package consul

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/metrics"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"git.qutoutiao.net/pedestal/discovery/util"
	"github.com/hashicorp/consul/api"
	"golang.org/x/sync/singleflight"
)

// adapter implements registry.Registrator interface by wrapping consul api.
type adapter struct {
	client  *api.Client
	options *option
	single  *singleflight.Group
	cache   *registry.ServiceCache

	watches         sync.Map
	watcher         registry.Watcher
	watchKeyChan    chan *watchKey
	watchResultChan chan *watchResult
	stopChan        chan struct{}
}

// New returns *adapter or error if failed.
func New(addr string, opts ...ConsulOption) (*adapter, error) {
	uri, err := url.Parse(addr)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	host, err := util.ResolveRandomHost(uri.Hostname())
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if uri.Port() != "" {
		host += ":" + uri.Port()
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	cfg.Transport.DialContext = WrapDialContext(host, cfg.Transport.DialContext)

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 默认设置
	options := &option{
		stale:             true,
		passingOnly:       true,
		threshold:         DefaultDegradeThreshold,
		emergency:         DefaultEmergencyThreshold,
		watchTimeout:      DefaultServiceWatchTimeout,
		watchLatency:      DefaultServiceWatchLatency,
		cacheSyncInterval: DefaultCacheSyncInterval,
		calmInterval:      DefaultCalmInterval,
	}
	for _, opt := range opts {
		opt(options)
	}

	consul := &adapter{
		options:         options,
		client:          client,
		single:          &singleflight.Group{},
		cache:           registry.NewServiceCache(options.cache, options.cacheSyncInterval),
		watchKeyChan:    make(chan *watchKey, 10),
		watchResultChan: make(chan *watchResult, 10),
		stopChan:        make(chan struct{}),
	}

	go consul.loop()

	return consul, nil
}

// Register tries to add a new service to consul catalog.
//
// NOTE: It will update service existed with the same service id!
func (ca *adapter) Register(svc *registry.Service, opts ...registry.RegistratorOption) error {
	o := new(registry.CommonRegistratorOption)
	for _, opt := range opts {
		switch ro := opt.(type) {
		case registry.RegisterOpt:
			ro(o)
		default:
			return errors.Wrap(errors.ErrArgument)
		}
	}

	if svc.Meta == nil {
		svc.Meta = make(map[string]string)
	}

	// Default metadata
	for k, v := range DefaultServiceMeta {
		if _, ok := svc.Meta[k]; !ok {
			svc.Meta[k] = v
		}
	}

	// Option metadata
	for k, v := range o.Metadata {
		svc.Meta[k] = v
	}

	service := &api.AgentServiceRegistration{
		ID:      svc.ServiceID(),
		Name:    svc.ServiceName(),
		Address: svc.ServiceIP(),
		Port:    svc.Port,
		Tags:    svc.Tags,
		Meta:    svc.Meta,
	}

	// adjust weight
	// NOTE: consul requires Weights.Passing > 0!
	weight := svc.ServiceWeight()
	if weight <= 0 {
		weight = registry.DefaultServiceWeight
	}

	service.Weights = &api.AgentWeights{
		Passing: weight,
		Warning: weight * 80 / 100,
	}

	service.Meta["weight"] = strconv.FormatInt(int64(weight), 10)

	// adjust checks
	for _, check := range o.Checks {
		tmpcheck, err := ca.normalizeCheck(service.Name, svc.Addr(), check)
		if err != nil {
			return errors.Wrap(err)
		}

		service.Checks = append(service.Checks, tmpcheck)
	}

	// adjust default tcp check for none custom check
	if len(service.Checks) <= 0 {
		service.Checks = append(service.Checks, &api.AgentServiceCheck{
			Name:                           service.Name,
			TCP:                            svc.Addr(),
			Status:                         api.HealthCritical,
			Interval:                       DefaultServiceCheckInterval,
			DeregisterCriticalServiceAfter: DefaultServiceDeregisterCriticalAfter,
		})
	}

	var (
		retries     uint32
		retryPeriod = 10 * time.Second
	)
	err := util.JitterTimeout(func() error {
		tmperr := ca.client.Agent().ServiceRegister(service)
		if tmperr != nil {
			logger.Errorf("consul.Register(%+v) with %v retries: %v", service, retries, tmperr)
		}

		atomic.AddUint32(&retries, 1)

		return tmperr
	}, retryPeriod, retryPeriod*DefaultRetryTimes, 2.0, true)

	if err != nil {
		return errors.Wrap(err)
	}

	logger.Infof("consul.Register(%+v): OK!", svc)
	return nil
}

// Deregister tries to remove the service with id from consul catalog.
//
// NOTE: It always returns true when trying to remove an un-exist service from consul catalog!
func (ca *adapter) Deregister(svc *registry.Service, opts ...registry.RegistratorOption) error {
	var (
		retries     uint32
		retryPeriod = 10 * time.Second
	)
	err := util.JitterTimeout(func() error {
		tmperr := ca.client.Agent().ServiceDeregister(svc.ServiceID())
		if tmperr != nil {
			logger.Errorf("consul.Deregister(%+v) with %d retries: %v", svc, retries, tmperr)
		}

		atomic.AddUint32(&retries, 1)

		return tmperr
	}, retryPeriod, retryPeriod*DefaultRetryTimes, 2.0, true)

	if err != nil {
		return errors.Wrap(err)
	}

	logger.Infof("consul.Deregister(%+v): Done!", svc)
	return nil
}

// GetServices implements registry.Discovery interface by resolving services from consul.
func (ca *adapter) GetServices(name string, opts ...registry.DiscoveryOption) ([]*registry.Service, error) {
	options := registry.NewCommonDiscoveryOption(opts...)
	key := registry.NewServiceKey(name, options.Tags, options.DC)

	// first, try cache
	services, err := ca.cache.GetServices(key)
	if err == nil && len(services) > 0 {
		return services, nil
	}

	entries, err, _ := ca.single.Do(key.ToString(), func() (interface{}, error) {
		var (
			services []*registry.Service
			err      error
		)
		if ca.options.useCatalog {
			services, err = ca.CatalogServices(name, options.DC, options.Tags)
		} else {
			services, err = ca.HealthServices(name, options.DC, options.Tags)
		}

		if err != nil {
			return nil, err
		}

		if len(services) == 0 {
			return nil, errors.ErrNotFound
		}

		metrics.GetMetrics().ReportTotalNodes("consul", key.Name, len(services))

		ca.cache.Set(key, services)

		// trigger watcher
		ca.watchKeyChan <- &watchKey{
			dc:   options.DC,
			name: name,
			tags: options.Tags,
		}

		return services, nil
	})

	if err != nil {
		return nil, err
	}

	if services, ok := entries.([]*registry.Service); ok {
		return services, nil
	}

	return nil, errors.ErrNotFound
}

// TODO: implements Notify later!
func (ca *adapter) Notify(event registry.Event) {}

// Watch sets callback of consul catalog.
func (ca *adapter) Watch(w registry.Watcher) {
	ca.watcher = w
}

func (ca *adapter) Close() {
	close(ca.stopChan)
}

func (ca *adapter) CatalogServices(name string, dc string, tags []string) ([]*registry.Service, error) {
	apiOpts := &api.QueryOptions{
		Datacenter: dc,
		AllowStale: ca.options.stale,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiOpts = apiOpts.WithContext(ctx)

	services, _, err := ca.client.Catalog().ServiceMultipleTags(name, tags, apiOpts)
	if err != nil {
		return nil, errors.Wrap(fmt.Errorf("consul.Catalog().HealthServices(%s, %v, %v, %+v): %v", name, tags, ca.options.passingOnly, apiOpts, err))
	}

	services = ReduceConsulCatalogWithoutMaint(services, ca.options.passingOnly)

	return ParseConsulCatalog(services), nil
}

func (ca *adapter) HealthServices(name string, dc string, tags []string) ([]*registry.Service, error) {
	apiOpts := &api.QueryOptions{
		Datacenter: dc,
		AllowStale: ca.options.stale,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiOpts = apiOpts.WithContext(ctx)

	services, _, err := ca.client.Health().ServiceMultipleTags(name, tags, ca.options.passingOnly, apiOpts)
	if err != nil {
		return nil, errors.Wrap(fmt.Errorf("consul.Health().HealthServices(%s, %v, %v, %+v): %v", name, tags, ca.options.passingOnly, apiOpts, err))
	}

	services = ReduceConsulHealthWithoutMaint(services)

	return ParseConsulHealth(services), nil
}

func (ca *adapter) normalizeCheck(name, addr string, check *registry.HealthCheck) (*api.AgentServiceCheck, error) {
	if check == nil {
		return nil, nil
	}

	interval := check.Interval
	if interval.Seconds() < MinServiceCheckInterval.Seconds() {
		interval = MinServiceCheckInterval
	}

	health := &api.AgentServiceCheck{
		Name:                           check.Name,
		Status:                         api.HealthCritical,
		Interval:                       interval.String(),
		DeregisterCriticalServiceAfter: DefaultServiceDeregisterCriticalAfter,
		Notes:                          name + "@" + addr,
	}

	switch check.Type {
	case registry.HealthTypeHTTP:
		health.HTTP = check.URI
		health.Method = check.Method
		health.Header = check.Header

		// for ip reuse cases
		if health.Header == nil {
			health.Header = map[string][]string{
				registry.HealthHTTPServiceNameKey: {name},
				registry.HealthHTTPServiceIPKey:   {addr},
			}
		} else {
			if _, ok := health.Header[registry.HealthHTTPServiceNameKey]; !ok {
				health.Header[registry.HealthHTTPServiceNameKey] = []string{name}
			}
			if _, ok := health.Header[registry.HealthHTTPServiceIPKey]; !ok {
				health.Header[registry.HealthHTTPServiceIPKey] = []string{addr}
			}
		}

	case registry.HealthTypeTCP:
		health.TCP = check.URI

	default:
		return nil, errors.Wrap(errors.ErrArgument)
	}

	return health, nil
}

func (ca *adapter) loop() {
	for {
		select {
		case key := <-ca.watchKeyChan:
			ca.startWatch(key)

		case result := <-ca.watchResultChan:
			ca.syncService(result, false)

		case <-ca.stopChan:
			ca.watches.Range(func(key, value interface{}) bool {
				if w, ok := value.(*Watch); ok && w != nil {
					logger.Infof("consul.Watch(%+v, %+v): stopped!", key, w)

					w.Stop()
				}

				return true
			})

			return
		}
	}
}

// syncService tries to update services of the key.
//
// NOTE: It's NOT thread safe!!!
func (ca *adapter) syncService(result *watchResult, overwrite bool) {
	key := registry.NewServiceKey(result.key.name, result.key.tags, result.key.dc)

	var services []*registry.Service
	if len(result.entries) <= 0 {
		if !overwrite {
			return
		}

		services = make([]*registry.Service, 0)
	} else {
		services = ParseConsulHealth(result.entries)
	}

	metrics.GetMetrics().ReportTotalNodes("consul", key.Name, len(services))

	ca.cache.Set(key, services)

	if ca.watcher != nil {
		ca.watcher.Handle(key, services)
	}
}

// TODO: refactor all watch with one goroutine!
func (ca *adapter) startWatch(wkey *watchKey) {
	key := registry.NewServiceKey(wkey.name, wkey.tags, wkey.dc)
	if _, ok := ca.watches.Load(key); ok {
		return
	}

	_, err, _ := ca.single.Do(key.ToString(), func() (v interface{}, err error) {
		watcher := NewWatch(ca.client, ca.options, wkey, ca.watchResultChan)

		ca.watches.Store(key, watcher)

		go watcher.Watch()

		return
	})
	if err != nil {
		logger.Errorf("consul.startWatch(%+v): %v", wkey, err)
	}
}
