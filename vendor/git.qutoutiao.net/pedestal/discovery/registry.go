package discovery

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/pedestal/discovery/cache"
	"git.qutoutiao.net/pedestal/discovery/consul"
	"git.qutoutiao.net/pedestal/discovery/eds"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/file"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/metrics"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"git.qutoutiao.net/pedestal/discovery/sentinel"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Registry wraps both register and discovery interfaces.
type Registry struct {
	lock     sync.Mutex
	opts     *registryOption
	differ   func(key registry.ServiceKey, services []*registry.Service) []*registry.Service
	watchers []registry.Watcher
	isClosed int32
}

// NewRegistry creates a new *Registry with given register or resolver implementation.
func NewRegistry(opts ...RegistryOption) (*Registry, error) {
	o := new(registryOption)
	for _, opt := range opts {
		opt(o)
	}

	r := &Registry{
		opts: o,
	}

	// apply watchers
	for _, adapter := range r.opts.discoveries {
		adapter.Watch(registry.WatchFunc(r.watchServices))
	}

	metrics.GetMetrics().ReportClientVersion(metrics.Version)

	return r, nil
}

// NewRegistryWithConsul creates a new *Registry with consul adapter as default discovery and register. It
// creates a consul.Cache with filepath.Join(os.TempDir(), "discovery-local") for local cache.
func NewRegistryWithConsul(addr string, opts ...consul.ConsulOption) (*Registry, error) {
	localDir := filepath.Join(os.TempDir(), DefaultTempDir)

	err := os.MkdirAll(localDir, 0777)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return NewRegistryWithConsulAndFile(addr, localDir, opts...)
}

// NewRegistryWithConsulAndFile creates a new *Registry with consul adapter as default discovery and register, and uses
// localDir as cache for local discovery.
// NOTE: It the customer' ability to ensure the fileDir is existed and can write!
func NewRegistryWithConsulAndFile(consulAddr, localDir string, opts ...consul.ConsulOption) (*Registry, error) {
	cacher, err := cache.New(cache.WithLocalDir(localDir), cache.WithFormat(cache.FormatDiscovery))
	if err != nil {
		return nil, errors.Wrap(err)
	}

	opts = append(opts, consul.WithCache(cacher))

	consulAdapter, consulErr := consul.New(consulAddr, opts...)
	if consulErr != nil {
		return nil, errors.Wrap(consulErr)
	}

	sentinelAdapter, sentinelErr := sentinel.New(DefaultSentinelAddr)
	if sentinelErr != nil {
		return nil, errors.Wrap(sentinelErr)
	}

	fileAdapter := file.New(cacher)

	return NewRegistry(WithDiscoveries(consulAdapter, sentinelAdapter, fileAdapter), WithRegisters(consulAdapter))
}

// NewRegistryWithConsulAndEDS creates a new *Registry with eds adapter as default discovery and consul adapter as default register.
func NewRegistryWithConsulAndEDS(consulAddr, edsAddr string, opts ...consul.ConsulOption) (*Registry, error) {
	consulAdapter, consulErr := consul.New(consulAddr, opts...)
	if consulErr != nil {
		return nil, errors.Wrap(consulErr)
	}

	edsAdapter, edsErr := eds.NewWithInterval(edsAddr, 10*time.Second)
	if edsErr != nil {
		return nil, errors.Wrap(edsErr)
	}

	sentinelAdapter, sentinelErr := sentinel.New(DefaultSentinelAddr)
	if sentinelErr != nil {
		return nil, errors.Wrap(sentinelErr)
	}

	return NewRegistry(WithDiscoveries(edsAdapter, consulAdapter, sentinelAdapter), WithRegisters(consulAdapter))
}

// LookupServices tries to resolve services of the name from registered discovery. It will retries among all discoveries
// when the result is unexpected.
func (r *Registry) LookupServices(name string, opts ...registry.DiscoveryOption) ([]*registry.Service, error) {
	if r.isClose() {
		return nil, errors.ErrClosed
	}

	if len(name) == 0 {
		return nil, errors.New("service name is required")
	}

	options := registry.NewCommonDiscoveryOption(opts...)
	key := registry.NewServiceKey(name, options.Tags, options.DC)

	var (
		services []*registry.Service
		err      error
	)
	for _, adapter := range r.opts.discoveries {
		tmpServices, tmpErr := adapter.GetServices(name, opts...)
		if tmpErr != nil {
			err = tmpErr

			// NOTE: always skip for error of not found!
			if errors.Is(tmpErr, errors.ErrNotFound) {
				logger.Warnf("%T.LookupServices(%s, %+v): %v", adapter, name, opts, tmpErr)
			} else {
				logger.Errorf("%T.LookupServices(%s, %+v): %+v", adapter, name, opts, tmpErr)
			}

			if r.opts.failType.IsFailFast() {
				return nil, errors.Wrap(tmpErr)
			}

			continue
		}

		if len(tmpServices) > len(services) {
			services = tmpServices
		}

		if r.isFallback(key, tmpServices) {
			adapter.Notify(registry.EventDegrade)
			continue
		}

		adapter.Notify(registry.EventRecover)

		return r.diffServices(key, services), nil
	}

	// now, we run into degrade for all discoveries!
	if len(services) > 0 {
		err = nil
	}

	return r.diffServices(key, services), err
}

// Register tries to register service with all registrators and returns wrapped service registrator which use for deregister service
// by one call.
func (r *Registry) Register(service *registry.Service, opts ...registry.RegistratorOption) (registrator ServiceRegister, err error) {
	if r.isClose() {
		return nil, errors.ErrClosed
	}

	err = service.Valid()
	if err != nil {
		return
	}

	for _, register := range r.opts.registrators {
		err = register.Register(service, opts...)
		if err != nil {
			logger.Errorf("%T.Register(%+v, %v): %+v", register, service, opts, err)

			if r.opts.failType == FailFast {
				return nil, errors.Wrap(err)
			}

			continue
		}
	}

	return &ServiceRegistrator{
		service:      service,
		registrators: r.opts.registrators,
	}, err
}

func (r *Registry) WithWatcher(w registry.Watcher) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.watchers = append(r.watchers, w)
}

func (r *Registry) WithWatcherFunc(key registry.ServiceKey, w registry.Watcher) {
	r.WithWatcher(registry.WatchFunc(func(yek registry.ServiceKey, services []*registry.Service) {
		if yek != key {
			return
		}

		w.Handle(yek, services)
	}))
}

func (r *Registry) WithDiffer(fn func(key registry.ServiceKey, services []*registry.Service) []*registry.Service) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.differ = fn
}

func (r *Registry) Close() {
	isSwap := atomic.CompareAndSwapInt32(&r.isClosed, 0, 1)
	if !isSwap {
		return
	}

	for _, discovery := range r.opts.discoveries {
		discovery.Close()
	}
}

func (r *Registry) isClose() bool {
	return atomic.LoadInt32(&r.isClosed) == 1
}

func (r *Registry) watchServices(key registry.ServiceKey, services []*registry.Service) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.isFallback(key, services) {
		logger.Errorf("Registry fallback triggered: key=%s, services=%d", key.ToString(), len(services))

		var err error

		services, err = r.LookupServices(key.Name, registry.WithDC(key.DC), registry.WithTags(strings.Split(key.Tags, ":")))
		if err != nil {
			logger.Errorf("registry.LookupServices(%s): %+v (fallback)", key.ToString(), err)
			return
		}
	}

	services = r.diffServices(key, services)

	for _, w := range r.watchers {
		w.Handle(key, services)
	}
}

func (r *Registry) diffServices(key registry.ServiceKey, services []*registry.Service) []*registry.Service {
	if r.differ == nil {
		return services
	}

	return r.differ(key, services)
}

func (r *Registry) isFallback(key registry.ServiceKey, services []*registry.Service) bool {
	// fallback with empty service
	if len(services) <= 0 {
		return true
	}

	// fallback with bootstrap strategy
	if r.opts.bootstrap[key] > len(services) {
		return true
	}

	return false
}
