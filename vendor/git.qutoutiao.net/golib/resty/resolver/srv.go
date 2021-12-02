package resolver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/pedestal/discovery/balancer/weightedroundrobin"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"git.qutoutiao.net/pedestal/discovery/util"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

// SRVResolver implements balancer.Interface interface for static ip list.
type SRVResolver struct {
	Protocol string
	Domain   string
	TTL      time.Duration
	Factor   float64

	store   sync.Map // for service key => watch handler
	looping chan bool
}

// TODO: filter records with given tags and datacenter?!
func (srv *SRVResolver) LookupServices(name string, options ...registry.DiscoveryOption) (services []*registry.Service, err error) {
	opts := registry.NewCommonDiscoveryOption(options...)

	_, records, err := net.LookupSRV(name, srv.Protocol, srv.Domain)
	if err != nil {
		return
	}

	services = make([]*registry.Service, len(records))
	for i, record := range records {
		services[i] = &registry.Service{
			ID:     uuid.New().String(),
			Name:   name,
			IP:     strings.TrimRight(record.Target, "."),
			Port:   int(record.Port),
			Weight: int32(record.Weight),
			Tags:   opts.Tags,
			Meta: map[string]string{
				"cloud":     "SRV",
				"container": "DNS",
				"dc":        opts.DC,
				"domain":    srv.Domain,
				"weight":    strconv.FormatUint(uint64(record.Weight), 10),
			},
		}
	}

	return
}

func (srv *SRVResolver) WithWatcherFunc(key registry.ServiceKey, watcher registry.Watcher) {
	srv.store.Store(key, watcher)
}

func (srv *SRVResolver) loop() {
	srv.looping = make(chan bool)

	for {
		select {
		case <-srv.looping:
			srv.looping = nil
			return

		default:
			time.Sleep(util.Jitter(srv.TTL, srv.Factor))

			var (
				totalKeys, failedKeys float64
			)
			srv.store.Range(func(key, value interface{}) bool {
				totalKeys++

				serviceKey, ok := key.(registry.ServiceKey)
				if !ok {
					srv.store.Delete(key)

					return true
				}

				services, err := srv.LookupServices(serviceKey.Name, registry.WithDC(serviceKey.DC), registry.WithTags(strings.Split(serviceKey.Tags, ":")))
				if err != nil {
					failedKeys++

					return true
				}

				watcher, ok := value.(registry.Watcher)
				if !ok {
					srv.store.Delete(key)

					return true
				}

				watcher.Handle(serviceKey, services)
				return true
			})

			srv.Factor = failedKeys / totalKeys
		}
	}
}

func (srv *SRVResolver) close() {
	if srv.looping == nil {
		return
	}

	close(srv.looping)

	return
}

// srvRecord struct holds the data to query the srv record for the
// following service.
type srvRecord struct {
	domain string
	ttl    time.Duration

	single   *singleflight.Group
	store    sync.Map
	resolver *SRVResolver
}

func NewSRVResolver(domain string, ttl time.Duration) Interface {
	if ttl <= 0 {
		ttl = time.Minute
	}

	srv := &srvRecord{
		domain: domain,
		ttl:    ttl,
		single: new(singleflight.Group),
		resolver: &SRVResolver{
			Protocol: "tcp",
			Domain:   domain,
			TTL:      ttl,
		},
	}

	go srv.loop()

	return srv
}

func (srv *srvRecord) Resolve(ctx context.Context, name string) (service *registry.Service, err error) {
	opts, err := balancer.OptionFromContext(ctx)
	if err != nil {
		opts = balancer.CustomOption{
			DC:   "",
			Tags: nil,
		}
	}

	key := registry.NewServiceKey(name, opts.Tags, opts.DC)

	iface, err, _ := srv.single.Do(key.ToString(), func() (value interface{}, err error) {
		value, ok := srv.store.Load(key)
		if ok {
			return
		}

		value = weightedroundrobin.New(srv.resolver, balancer.WithDC(opts.DC), balancer.WithTags(opts.Tags...))

		srv.store.Store(key, value)
		return
	})
	if err != nil {
		return
	}

	lb, ok := iface.(balancer.Balancer)
	if !ok {
		err = fmt.Errorf("invalid loadbalancer(%T) of %v", iface, key)
		return
	}

	return lb.Next(ctx, name)
}

func (srv *srvRecord) Block(ctx context.Context, name string, service *registry.Service) {
}

func (srv *srvRecord) Close() {
	if srv.resolver == nil {
		return
	}

	srv.resolver.close()

	return
}

func (srv *srvRecord) loop() {
	if srv.resolver == nil {
		return
	}

	srv.resolver.loop()
}
