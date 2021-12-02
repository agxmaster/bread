package statics

import (
	"git.qutoutiao.net/pedestal/discovery/registry"
)

type builder struct {
	store map[registry.ServiceKey][]*registry.Service
}

func NewBuilder() *builder {
	return &builder{
		store: make(map[registry.ServiceKey][]*registry.Service),
	}
}

func (b *builder) Build() map[registry.ServiceKey][]*registry.Service {
	return b.store
}

func (b *builder) WithService(name string, services ...*registry.Service) *builder {
	return b.WithServiceAndTagsDC(name, nil, "", services...)
}

func (b *builder) WithServiceAndTags(name string, tags []string, services ...*registry.Service) *builder {
	return b.WithServiceAndTagsDC(name, tags, "", services...)
}

func (b *builder) WithServiceAndDC(name, dc string, services ...*registry.Service) *builder {
	return b.WithServiceAndTagsDC(name, nil, dc, services...)
}

func (b *builder) WithServiceAndTagsDC(name string, tags []string, dc string, services ...*registry.Service) *builder {
	key := registry.NewServiceKey(name, tags, dc)

	return b.WithServiceKey(key, services...)
}

func (b *builder) WithServiceKey(key registry.ServiceKey, services ...*registry.Service) *builder {
	b.store[key] = append(b.store[key], services...)

	return b
}
