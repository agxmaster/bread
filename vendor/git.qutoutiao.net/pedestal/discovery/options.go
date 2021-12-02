package discovery

import (
	"git.qutoutiao.net/pedestal/discovery/registry"
)

type RegistryOption func(o *registryOption)

type registryOption struct {
	registrators []registry.Registrator
	discoveries  []registry.Discovery
	bootstrap    map[registry.ServiceKey]int
	failType     FailType
}

// WithFailType configs registry with fail type
func WithFailType(t FailType) RegistryOption {
	return func(o *registryOption) {
		o.failType = t
	}
}

// WithRegisters configs registry with registrators
func WithRegisters(registrators ...registry.Registrator) RegistryOption {
	return func(o *registryOption) {
		o.registrators = append(o.registrators, registrators...)
	}
}

// WithDiscoveries configs registry with discoveries
func WithDiscoveries(discoveries ...registry.Discovery) RegistryOption {
	return func(o *registryOption) {
		o.discoveries = append(o.discoveries, discoveries...)
	}
}

// WithBootstrap configs registry with bootstrap
func WithBootstrap(bootstraps map[registry.ServiceKey]int) RegistryOption {
	return func(o *registryOption) {
		for key, expected := range bootstraps {
			WithBootstrapByKey(key, expected)
		}
	}
}

func WithBootstrapByKey(key registry.ServiceKey, expected int) RegistryOption {
	return func(o *registryOption) {
		if o.bootstrap == nil {
			o.bootstrap = make(map[registry.ServiceKey]int)
		}

		o.bootstrap[key] = expected
	}
}

func WithBootstrapByName(name string, expected int) RegistryOption {
	return WithBootstrapByKey(registry.NewServiceKey(name, nil, ""), expected)
}
