package resolver

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/pedestal/discovery/balancer/bwlist"
	"git.qutoutiao.net/pedestal/discovery/balancer/weightedroundrobin"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"golang.org/x/sync/singleflight"
)

var (
	DefaultConsulAddr = "http://127.0.0.1:8500"
)

// registryRecord struct holds the data to query the nodes for the
// following service.
type registryRecord struct {
	resolver      *discovery.Registry
	resolverValue *atomic.Value
	single        *singleflight.Group
	store         sync.Map
}

func NewRegistryResolver(resolver *discovery.Registry, err error) Interface {
	resolverValue := new(atomic.Value)
	if err != nil {
		resolverValue.Store(err)
	}

	record := &registryRecord{
		resolver:      resolver,
		resolverValue: resolverValue,
		single:        new(singleflight.Group),
	}

	return record
}

func NewConsulResolver(addr string) Interface {
	if len(addr) == 0 {
		addr = DefaultConsulAddr
	}

	return NewRegistryResolver(discovery.NewRegistryWithConsul(addr))
}

func NewEDSResolver(consulAddr, edsAddr string) Interface {
	if len(consulAddr) == 0 {
		consulAddr = DefaultConsulAddr
	}

	return NewRegistryResolver(discovery.NewRegistryWithConsulAndEDS(consulAddr, edsAddr))
}

func (record *registryRecord) Resolve(ctx context.Context, name string) (service *registry.Service, err error) {
	err = ctx.Err()
	if err != nil {
		return
	}

	err, ok := record.resolverValue.Load().(error)
	if ok && err != nil {
		return
	}

	opts, err := balancer.OptionFromContext(ctx)
	if err != nil {
		opts = balancer.CustomOption{
			DC:   "",
			Tags: nil,
		}
	}

	key := registry.NewServiceKey(name, opts.Tags, opts.DC)

	iface, err, _ := record.single.Do(key.ToString(), func() (value interface{}, err error) {
		value, ok := record.store.Load(key)
		if ok {
			return
		}

		value = bwlist.New(weightedroundrobin.New(record.resolver, balancer.WithDC(opts.DC), balancer.WithTags(opts.Tags...)))

		record.store.Store(key, value)
		return
	})
	if err != nil {
		return
	}

	lb, ok := iface.(balancer.Balancer)
	if !ok {
		err = fmt.Errorf("invalid loadbalancer(%T) of %+v", iface, key)
		return
	}

	return lb.Next(ctx, name)
}

func (record *registryRecord) Block(ctx context.Context, name string, service *registry.Service) {
	opts, err := balancer.OptionFromContext(ctx)
	if err != nil {
		opts = balancer.CustomOption{
			DC:   "",
			Tags: nil,
		}
	}

	key := registry.NewServiceKey(name, opts.Tags, opts.DC)

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

func (record *registryRecord) Close() {
	if record.resolver == nil {
		return
	}

	record.resolver.Close()
	record.resolver = nil
}
