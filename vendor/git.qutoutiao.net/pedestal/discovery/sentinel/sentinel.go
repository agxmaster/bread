package sentinel

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"git.qutoutiao.net/pedestal/discovery/consul"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/metrics"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"git.qutoutiao.net/pedestal/discovery/util"
	"golang.org/x/sync/singleflight"
)

type adapter struct {
	client  *http.Client
	addr    string
	opts    *option
	cache   *registry.ServiceCache
	watcher registry.Watcher

	//关注的服务
	keys   sync.Map
	cond   *sync.Cond
	single *singleflight.Group
	stopC  chan struct{}
}

func New(addr string, opts ...Option) (*adapter, error) {
	options := &option{
		cleanInterval: 1 * time.Hour,
		fetchInterval: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(options)
	}

	maxcpus := runtime.GOMAXPROCS(0)

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 5 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          maxcpus * 10,
			IdleConnTimeout:       15 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   maxcpus,
		}}

	sentinel := &adapter{
		client: client,
		cache:  registry.NewServiceCache(nil, 0),
		single: &singleflight.Group{},
		addr:   addr,
		opts:   options,
		cond:   sync.NewCond(&sync.Mutex{}),
		stopC:  make(chan struct{}),
	}

	go sentinel.loop()

	return sentinel, nil
}

func (sa *adapter) GetServices(name string, opts ...registry.DiscoveryOption) ([]*registry.Service, error) {
	options := registry.NewCommonDiscoveryOption(opts...)
	key := registry.NewServiceKey(name, options.Tags, options.DC)

	// first, try load from cache
	wkey := sa.loadWatchKey(key)
	if wkey != nil {
		if !wkey.useSentinel() {
			return nil, errors.ErrNotFound
		}

		wkey.setLastTime(time.Now())

		return sa.cache.GetServices(key)
	}

	_, err, _ := sa.single.Do(key.ToString(), func() (interface{}, error) {
		wkey := &watchKey{
			name: name,
			dc:   options.DC,
			tags: options.Tags,
			last: time.Now(),
		}

		sa.storeWatchKey(key, wkey)
		sa.cond.Broadcast()

		err := sa.fetchServices(wkey)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	services, err := sa.cache.GetServices(key)
	if err != nil {
		return nil, err
	}

	if len(services) <= 0 {
		return nil, errors.ErrNotFound
	}

	return services, nil
}

func (sa *adapter) Notify(event registry.Event) {}

func (sa *adapter) Watch(w registry.Watcher) {
	sa.watcher = w
}

func (sa *adapter) Close() {
	close(sa.stopC)
}

func (sa *adapter) loop() {
	timer := time.NewTimer(sa.opts.fetchInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			expiredKeys := make([]*watchKey, 0)

			sa.keys.Range(func(k interface{}, v interface{}) bool {
				if key, ok := v.(*watchKey); ok {
					if key.isExpired(sa.opts.cleanInterval) {
						expiredKeys = append(expiredKeys, key)
						return true
					}
				}

				sa.keys.Delete(k)
				return true
			})

			sa.cond.L.Lock()
			if len(expiredKeys) <= 0 {
				// block for available keys
				sa.cond.Wait()
			}
			sa.cond.L.Unlock()

			for _, key := range expiredKeys {
				time.Sleep(util.SlidingDuration(1 * time.Second))

				if err := sa.fetchServices(key); err != nil {
					if errors.Is(err, errors.ErrNotFound) {
						logger.Warnf("sentinel.fetchServices(%+v): %v", key, err)
					} else {
						logger.Errorf("sentinel.fetchServices(%+v): %+v", key, err)
					}
				}
			}

			timer.Reset(sa.opts.fetchInterval)

		case <-sa.stopC:
			return
		}
	}
}

func (sa *adapter) fetchServices(wkey *watchKey) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	params := make(url.Values)
	params.Add("dc", wkey.dc)
	for _, tag := range wkey.tags {
		params.Add("tag", tag)
	}

	uri := sa.addr + "/v1/snapshot/service/" + wkey.name + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	resp, err := sa.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// nolint: errcheck
		io.Copy(ioutil.Discard, resp.Body)

		return errors.Errorf("http.Get(%s) with invalid status code: %v", uri, resp.StatusCode)
	}

	data := new(SnapshotResponse)
	result := &Response{
		Data: data,
	}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return err
	}

	wkey.setSentinel(data.UseSentinel)
	if !wkey.useSentinel() {
		return nil
	}

	if len(data.Data) <= 0 {
		return nil
	}

	key := registry.NewServiceKey(wkey.name, wkey.tags, wkey.dc)
	services := consul.ParseConsulCatalog(consul.ReduceConsulCatalogWithoutMaint(data.Data, false))

	metrics.GetMetrics().ReportTotalNodes("sentinel", key.Name, len(services))

	sa.cache.Set(key, services)

	if sa.watcher != nil {
		sa.watcher.Handle(key, services)
	}

	return nil
}

func (sa *adapter) loadWatchKey(key registry.ServiceKey) *watchKey {
	v, ok := sa.keys.Load(key)
	if !ok {
		return nil
	}

	if wkey, ok := v.(*watchKey); ok {
		return wkey
	}

	return nil
}

func (sa *adapter) storeWatchKey(key registry.ServiceKey, data *watchKey) {
	sa.keys.Store(key, data)
}
