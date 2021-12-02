package balancer

import (
	"strconv"
	"sync"

	"git.qutoutiao.net/gopher/qms/internal/core/registry"
)

var (
	r    Resolver
	once sync.Once
)

type Resolver interface {
	Lookup(key *registry.ServiceKey) ([]*Service, error)
	WithWatchFunc(key *registry.ServiceKey, watcher Watcher)
}

// call_back机制
type WatchFunc func(key *registry.ServiceKey, services []*Service)

func (f WatchFunc) Watch(key *registry.ServiceKey, services []*Service) {
	f(key, services)
}

type Watcher interface {
	Watch(key *registry.ServiceKey, instances []*Service)
}

type resolver struct {
	discovery registry.Registry
}

func (r *resolver) Lookup(key *registry.ServiceKey) ([]*Service, error) {
	instances, err := r.discovery.Lookup(key)
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0, len(instances))
	for _, v := range instances {
		services = append(services, &Service{
			ID:       v.ID,
			Name:     v.Name,
			IP:       v.IP,
			Port:     v.Port,
			Endpoint: v.IP + ":" + strconv.Itoa(v.Port),
			Weight:   v.Weight,
			Tags:     v.Tags,
			Meta:     v.Meta,
		})
	}
	return services, nil
}

func (r *resolver) WithWatchFunc(key *registry.ServiceKey, w Watcher) {
	r.discovery.Watch(key, registry.WatchFunc(func(yek *registry.ServiceKey, instances []*registry.Service) {
		// 底层已经对key做过校验 此处不需要二次校验
		services := make([]*Service, 0, len(instances))
		for _, v := range instances {
			services = append(services, &Service{
				ID:       v.ID,
				Name:     v.Name,
				IP:       v.IP,
				Port:     v.Port,
				Endpoint: v.IP + ":" + strconv.Itoa(v.Port),
				Weight:   v.Weight,
				Tags:     v.Tags,
				Meta:     v.Meta,
			})
		}
		w.Watch(key, services)
	}))
}

func getResolver() Resolver {
	once.Do(func() {
		r = &resolver{
			discovery: registry.GetRegistry(),
		}
	})
	return r
}
