package file

import (
	"sync"

	"git.qutoutiao.net/pedestal/discovery/cache"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/metrics"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"git.qutoutiao.net/pedestal/discovery/statics"
)

// File implements registry.Discovery interface with local file.
type File struct {
	mux     sync.Mutex
	options *option
	loader  Loader
	watcher registry.Watcher
	store   sync.Map
}

func New(cacher cache.Interface, opts ...Option) *File {
	return NewWithLoader(cacher, opts...)
}

func NewWithLoader(loader Loader, opts ...Option) *File {
	options := new(option)
	for _, opt := range opts {
		opt(options)
	}

	return &File{
		options: options,
		loader:  loader,
	}
}

func (f *File) GetServices(name string, opts ...registry.DiscoveryOption) ([]*registry.Service, error) {
	options := registry.NewCommonDiscoveryOption(opts...)
	key := registry.NewServiceKey(name, options.Tags, options.DC)

	adapter, err := f.loadAdapter(key)
	if err != nil {
		return nil, err
	}

	services, err := adapter.GetServices(name, opts...)
	if err != nil {
		return nil, err
	}

	metrics.GetMetrics().ReportTotalNodes("file", name, len(services))
	return services, err
}

func (f *File) Watch(w registry.Watcher) {
	f.watcher = w
}

func (f *File) Notify(event registry.Event) {}

func (f *File) Close() {}

func (f *File) loadAdapter(key registry.ServiceKey) (registry.Discovery, error) {
	f.mux.Lock()
	defer f.mux.Unlock()

	iface, ok := f.store.Load(key)
	if !ok {
		iface = statics.NewWithLoader(f.loader)

		f.store.Store(key, iface)
	}

	d, ok := iface.(registry.Discovery)
	if !ok {
		return nil, errors.ErrNotFound
	}

	return d, nil
}
