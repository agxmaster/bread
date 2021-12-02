package weightedroundrobin

import (
	"context"
	"sync"

	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"golang.org/x/sync/singleflight"
)

// New returns a weighted round-robin balancer.Balancer with balancer.Resolver given.
func New(resolver balancer.Resolver, opts ...balancer.Option) balancer.Balancer {
	o := new(balancer.CustomOption)
	for _, opt := range opts {
		opt(o)
	}

	lb := &weightedRoundRobin{
		opts:     o,
		resolver: resolver,
		single:   new(singleflight.Group),
	}

	return lb
}

// weightedRoundRobin implements balancer.Balancer interface with weighted round-robin alg.
type weightedRoundRobin struct {
	opts     *balancer.CustomOption
	resolver balancer.Resolver
	store    sync.Map
	single   *singleflight.Group
}

func (wrr *weightedRoundRobin) Next(ctx context.Context, name string) (*registry.Service, error) {
	opt, err := balancer.OptionFromContext(ctx)
	if err != nil {
		opt = *wrr.opts
	}

	key := registry.NewServiceKey(name, opt.Tags, opt.DC)

	iface, ok := wrr.store.Load(key)
	if ok {
		return iface.(*wrrLoadBalance).next()
	}

	iface, err, _ = wrr.single.Do(key.ToString(), func() (interface{}, error) {
		services, err := wrr.resolver.LookupServices(name, registry.WithDC(opt.DC), registry.WithTags(opt.Tags))
		if err != nil {
			return nil, errors.Wrap(err)
		}

		if len(services) <= 0 {
			return nil, errors.Wrap(errors.ErrNotFound)
		}

		tmplb := &wrrLoadBalance{}
		tmplb.update(services)

		wrr.resolver.WithWatcherFunc(key, wrr)
		wrr.store.Store(key, tmplb)

		return tmplb, nil
	})
	if err != nil {
		return nil, err
	}

	if lb, ok := iface.(*wrrLoadBalance); ok {
		return lb.next()
	}

	return nil, errors.ErrArgument
}

func (wrr *weightedRoundRobin) Handle(key registry.ServiceKey, services []*registry.Service) {
	// void pollution with invalid data
	if len(services) <= 0 {
		return
	}

	iface, ok := wrr.store.Load(key)
	if ok {
		iface.(*wrrLoadBalance).update(services)
		return
	}

	lb := &wrrLoadBalance{}
	lb.update(services)

	wrr.store.Store(key, lb)
	return
}
