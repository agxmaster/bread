package resolver

import (
	"context"

	"git.qutoutiao.net/pedestal/discovery/registry"
)

type Interface interface {
	Resolve(ctx context.Context, name string) (service *registry.Service, err error)
	Block(ctx context.Context, name string, service *registry.Service)
	Close()
}
