package balancer

import (
	"context"

	"git.qutoutiao.net/pedestal/discovery/registry"
)

// Balancer represents balancer interface.
type Balancer interface {
	Next(ctx context.Context, name string) (*registry.Service, error)
}

// Resolver represents resolver interface for custom balancer.
type Resolver interface {
	WithWatcherFunc(key registry.ServiceKey, watcher registry.Watcher)
	LookupServices(name string, opts ...registry.DiscoveryOption) (services []*registry.Service, err error)
}

type BWLister interface {
	Block(service *registry.Service)
	Unblock(service *registry.Service)
}
