package resolver

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/pedestal/discovery/balancer/bwlist"
	"git.qutoutiao.net/pedestal/discovery/balancer/weightedroundrobin"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

// staticRecord struct holds the data to query the static record for the
// following service, such as domain, ipv4:port, etc.
type staticRecord struct {
	resolver *RecordResolver
	single   *singleflight.Group
	store    sync.Map
}

func NewStaticResolver(records map[string][]string) Interface {
	srv := &staticRecord{
		single:   new(singleflight.Group),
		resolver: NewRecordResolver(records),
	}

	return srv
}

func (record *staticRecord) Resolve(ctx context.Context, name string) (service *registry.Service, err error) {
	key := registry.NewServiceKey(name, nil, "")

	iface, err, _ := record.single.Do(key.ToString(), func() (value interface{}, err error) {
		value, ok := record.store.Load(key)
		if ok {
			return
		}

		value = bwlist.New(weightedroundrobin.New(record.resolver))

		record.store.Store(key, value)
		return
	})
	if err != nil {
		return
	}

	lb, ok := iface.(balancer.Balancer)
	if !ok {
		err = fmt.Errorf("invalid load balancer(%T) of %v", iface, key)
		return
	}

	return lb.Next(ctx, name)
}

func (record *staticRecord) Block(ctx context.Context, name string, service *registry.Service) {
	key := registry.NewServiceKey(name, nil, "")

	value, ok := record.store.Load(key)
	if !ok {
		return
	}

	bwl, ok := value.(balancer.BWLister)
	if !ok {
		return
	}

	bwl.Block(service)
}

func (record *staticRecord) Close() {
	if record.resolver == nil {
		return
	}

	record.resolver.close()
	record.resolver = nil
}

// RecordResolver implements balancer.Interface interface for static ip list.
type RecordResolver struct {
	store sync.Map // for service key => watch handler
}

func NewRecordResolver(staticRecords map[string][]string) *RecordResolver {
	static := &RecordResolver{}

	for name, records := range staticRecords {
		static.init(name, records)
	}

	return static
}

func (static *RecordResolver) init(service string, records []string) {
	key := registry.NewServiceKey(service, nil, "")

	var services []*registry.Service
	for _, record := range records {
		if len(record) == 0 {
			continue
		}

		urlobj, err := url.Parse(record)
		if err != nil {
			ip2port := strings.SplitN(record, ":", 2)
			if net.ParseIP(ip2port[0]) == nil {
				err = fmt.Errorf("invalid record: %s", record)
			} else {
				err = nil

				urlobj = &url.URL{
					Host: record,
				}
			}
		}
		if err != nil {
			log.Println(err.Error())
			continue
		}

		// adjust hostname for example.com format
		hostname := urlobj.Hostname()
		if len(hostname) == 0 {
			switch {
			case len(urlobj.Scheme) > 0: // for www.example.com:80/api format
				hostname = urlobj.Scheme

				if len(urlobj.Opaque) > 0 {
					urlobj.Host = hostname + ":" + strings.SplitN(urlobj.Opaque, "/", 2)[0]
				}

			case len(urlobj.Path) > 0: // for for www.example.com/api format
				hostname = strings.SplitN(urlobj.Path, "/", 2)[0]

			}

			if len(hostname) == 0 {
				hostname = record
			}
		}

		port := 80
		switch urlobj.Scheme {
		case "https":
			port = 443
		}
		if urlport := urlobj.Port(); len(urlport) > 0 {
			n, err := strconv.Atoi(urlport)
			if err == nil {
				port = n
			}
		}

		services = append(services, &registry.Service{
			ID:     uuid.New().String(),
			Name:   service,
			IP:     hostname,
			Port:   port,
			Weight: 100,
			Meta: map[string]string{
				"cloud":  "static",
				"weight": "100",
			},
		})
	}

	static.store.Store(key, services)
}

// TODO: filter records with given tags and datacenter?!
func (static *RecordResolver) LookupServices(name string, options ...registry.DiscoveryOption) (services []*registry.Service, err error) {
	key := registry.NewServiceKey(name, nil, "")

	iface, ok := static.store.Load(key)
	if !ok {
		err = errors.ErrNotFound
		return
	}

	services, ok = iface.([]*registry.Service)
	if !ok {
		err = errors.ErrInvalidService

		static.store.Delete(key)
	}

	return
}

func (static *RecordResolver) WithWatcherFunc(key registry.ServiceKey, watcher registry.Watcher) {
	services, err := static.LookupServices(key.Name)
	if err != nil {
		watcher.Handle(key, nil)
	} else {
		watcher.Handle(key, services)
	}
}

func (static *RecordResolver) close() {
	static.store.Range(func(key, value interface{}) bool {
		static.store.Delete(key)
		return true
	})

	return
}
