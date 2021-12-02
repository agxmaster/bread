package statics

import (
	"sync"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

type storedService struct {
	services []*registry.Service
	err      error
}

// Statics implements registry.Discovery interface for dns, service list.
type Statics struct {
	mux    sync.RWMutex
	loader Loader
	store  map[registry.ServiceKey]*storedService
	once   sync.Once
}

func New(list map[registry.ServiceKey][]*registry.Service) *Statics {
	loader := NewStaticLoader(list)

	return NewWithLoader(loader)
}

// see builder.go, helper function to build with static loader
func NewWithBuilder(build *builder) *Statics {
	services := build.Build()

	return New(services)
}

// NewWithLoader creates an static adapter with loader given.
func NewWithLoader(loader Loader) *Statics {
	return &Statics{
		loader: loader,
	}
}

func (s *Statics) GetServices(name string, opts ...registry.DiscoveryOption) ([]*registry.Service, error) {
	s.once.Do(func() {
		s.store = make(map[registry.ServiceKey]*storedService)
	})

	o := registry.NewCommonDiscoveryOption(opts...)

	key := registry.NewServiceKey(name, o.Tags, o.DC)

	// first, try stored from memory.
	s.mux.RLock()
	stored, ok := s.store[key]
	if ok {
		s.mux.RUnlock()

		if stored.err != nil {
			return nil, stored.err
		}

		return stored.services, nil
	}
	s.mux.RUnlock()

	// second, load from Loader
	s.mux.Lock()
	defer s.mux.Unlock()

	services, err := s.loader.Load(key)
	if err != nil {
		s.store[key] = &storedService{
			services: services,
			err:      err,
		}

		return nil, err
	}

	if len(services) == 0 {
		err := errors.Wrap(errors.ErrNotFound)

		s.store[key] = &storedService{
			services: nil,
			err:      err,
		}

		return nil, err
	}

	s.store[key] = &storedService{
		services: services,
		err:      nil,
	}

	return services, nil
}

func (s *Statics) Watch(w registry.Watcher) {}

func (s *Statics) Notify(event registry.Event) {}

func (s *Statics) Close() {}
