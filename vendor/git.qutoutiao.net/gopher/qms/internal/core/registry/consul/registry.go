package consul

import (
	"strconv"

	"git.qutoutiao.net/gopher/qms/internal/core/registry"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/registryutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/discovery/logger"
	qregistry "git.qutoutiao.net/pedestal/discovery/registry"
)

const Name = "consul"

type registryBuilder struct{}

func (r *registryBuilder) Build(config *registry.Config) (registry.Registry, error) {
	var (
		remote *discovery.Registry
		err    error
		local  = make(map[string][]*registry.Service)
	)

	if config.Local != nil {
		local = config.Local
	}
	if config.Pilot != "" {
		remote, err = discovery.NewRegistryWithConsulAndEDS(config.Consul, config.Pilot)
	} else {
		remote, err = discovery.NewRegistryWithConsulAndFile(config.Consul, config.CacheDir)
	}
	if err != nil {
		return nil, errors.Wrapf(ErrNewRegistry, "consul: %s, pilot: %s, err: %s", config.Consul, config.Pilot, err.Error())
	}
	return &Registry{
		config:      config,
		local:       local,
		remote:      remote,
		deregisters: make(map[protocol.Protocol][]Deregister),
	}, nil
}

type Registry struct {
	config      *registry.Config
	local       map[string][]*registry.Service // map[name][]service
	remote      *discovery.Registry
	deregisters map[protocol.Protocol][]Deregister // 临时方案，后续还是要通过Name来Map
}

func (r *Registry) Register(service *registry.Service, option ...registry.RegisterOption) error {
	var ops []qregistry.RegistratorOption
	if hc := service.HC; hc != nil {
		ops = append(ops, qregistry.WithHTTPHealthCheck(&qregistry.HTTPHealthCheck{
			Name:     registryutil.Healthcheck,
			URI:      hc.HTTP,
			Interval: registryutil.DefaultHCInterval,
			Method:   hc.Method,
			Header:   hc.Header,
		}))
	}

	deregister, err := r.remote.Register(&qregistry.Service{
		Name:   service.Name,
		IP:     service.IP,
		Port:   service.Port,
		Weight: service.Weight,
		Tags:   service.Tags,
		Meta:   service.Meta,
	}, ops...)
	if err != nil {
		return errors.Wrapf(ErrRegister, "name: %s, endpoint: %s, err: %v", service.Name, service.Endpoint, err)
	}
	r.deregisters[service.Protocol] = append(r.deregisters[service.Protocol], newDeregister(service.Name, deregister))
	return nil
}

// TODO: Deregister 后续改为服务名为key的形式，Service打通之后
func (r *Registry) Deregister(proto protocol.Protocol, option ...registry.RegisterOption) {
	for _, deregister := range r.deregisters[proto] {
		if err := deregister.Deregister(); err != nil {
			qlog.WithError(err).Errorf("service: %s, err: %v", deregister.Name(), err)
			continue
		}
		qlog.Infof("deregister service[%s] success.", deregister.Name())
	}
}

func (r *Registry) Lookup(key *registry.ServiceKey) ([]*registry.Service, error) {
	// lookup from local
	if services, ok := r.local[key.Name()]; ok {
		return services, nil
	}

	// lookup from remote
	options := make([]qregistry.DiscoveryOption, 0, 2)
	if key.DC() != "" {
		options = append(options, qregistry.WithDC(key.DC()))
	}
	if len(key.Tags()) > 0 {
		options = append(options, qregistry.WithTags(key.Tags()))
	}
	services, err := r.remote.LookupServices(key.Name(), options...)
	if err != nil {
		return nil, errors.Wrapf(registry.ErrNotFoundService, "lookup services error: %v", err)
	}

	// TODO: DRY
	result := make([]*registry.Service, 0, len(services))
	for _, v := range services {
		result = append(result, &registry.Service{
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
	return result, nil
}

func (r *Registry) Watch(key *registry.ServiceKey, watcher registry.Watcher) {
	qkey := qregistry.NewServiceKey(key.Name(), key.Tags(), key.DC())

	r.remote.WithWatcher(qregistry.WatchFunc(func(yek qregistry.ServiceKey, services []*qregistry.Service) {
		if qkey != yek {
			return
		}
		if len(services) == 0 {
			return
		}

		// TODO: DRY
		result := make([]*registry.Service, 0, len(services))
		for _, v := range services {
			result = append(result, &registry.Service{
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
		watcher.Handler(key, result)
	}))
}

func (r *Registry) Close() error {
	r.remote.Close()
	return nil
}

func init() {
	logger.SetLogger(qlog.GetLogger())
	registry.RegisterBuilder(Name, &registryBuilder{})
}
