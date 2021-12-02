package transport

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/loader"
	"git.qutoutiao.net/golib/resty/logger"
	"git.qutoutiao.net/golib/resty/metrics"
	"git.qutoutiao.net/golib/resty/resolver"
	"git.qutoutiao.net/golib/resty/trace"
	"git.qutoutiao.net/pedestal/discovery"
	"github.com/prometheus/client_golang/prometheus"
)

type ChainTransport struct {
	tripper      http.RoundTripper
	value        atomic.Value
	looping      chan struct{}
	loader       loader.Interface
	logger       logger.Interface
	transform4xx bool
}

type configValue struct {
	connects    sync.Map // for global shared connects
	services    sync.Map // for custom services
	subconnects sync.Map // for custom service connect, available values are [sidecar|consul|static]
}

func (value *configValue) HealthCheck() {
	if value == nil {
		return
	}

	value.connects.Range(func(key, value interface{}) bool {
		tripper, ok := value.(Interface)
		if ok {
			tripper.Check()
		}

		return true
	})
	value.subconnects.Range(func(key, value interface{}) bool {
		tripper, ok := value.(Interface)
		if ok {
			tripper.Check()
		}

		return true
	})
}

func NewChainTransport(tripper http.RoundTripper) *ChainTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	chain := &ChainTransport{
		tripper: tripper,
		logger:  logger.NewWithPrefix("Resty"),
	}
	go chain.loop()

	return chain
}

func NewChainTransportFromConfig(cfg *config.Config, tripper http.RoundTripper) (chain *ChainTransport, err error) {
	if cfg == nil {
		err = fmt.Errorf("invalid config")
		return
	}

	chain = NewChainTransport(tripper)
	chain.ReloadConfig(cfg)

	// apply loader if defined
	if cfg.Loader.IsValid() {
		cloader, err := loader.New(cfg.Loader)
		if err == nil {
			chain.WithLoader(cloader)
		}
	}

	return
}

func (chain *ChainTransport) WithProxy(proxyURL *url.URL) *ChainTransport {
	transport, ok := chain.tripper.(*http.Transport)
	if !ok {
		return chain
	}

	if proxyURL == nil {
		transport.Proxy = nil

		return chain
	}

	transport.Proxy = http.ProxyURL(proxyURL)
	return chain
}

func (chain *ChainTransport) WithLoader(loader loader.Interface) *ChainTransport {
	if loader == nil {
		return chain
	}

	if chain.loader != nil {
		chain.loader.Stop()
	}

	chain.loader = loader

	// setup chain with config loaded
	cfg := chain.loader.GetConfig()
	if cfg != nil {
		chain.ReloadConfig(cfg)
	}

	chain.loader.WithHandler(chain.ReloadConfig)

	return chain
}

func (chain *ChainTransport) WithTransform4xxError(enable bool) *ChainTransport {
	chain.transform4xx = enable
	return chain
}

func (chain *ChainTransport) WithTLSClientConfig(config *tls.Config) *ChainTransport {
	transport, ok := chain.tripper.(*http.Transport)
	if ok {
		transport.TLSClientConfig = config
	}

	return chain
}

func (chain *ChainTransport) WithServiceConfig(services ...*config.ServiceConfig) *ChainTransport {
	if len(services) == 0 {
		return chain
	}

	value := chain.load()
	for _, service := range services {
		if service == nil {
			continue
		}

		value.services.Store(service.Name, service)

		if connector := chain.newConnector(service.Name, service.Connect); connector != nil {
			value.subconnects.Store(service.Name, connector)
		}
	}

	chain.value.Store(value)

	return chain
}

func (chain *ChainTransport) CleanServiceConfig(services ...*config.ServiceConfig) {
	if len(services) == 0 {
		return
	}

	value := chain.load()
	for _, service := range services {
		if service == nil {
			continue
		}

		value.services.Delete(service.Name)
		value.subconnects.Delete(service.Name)
	}

	chain.value.Store(value)
}

func (chain *ChainTransport) WithStatic(records map[string][]string) *ChainTransport {
	chain.addConnect(ProviderStatic, NewStaticTransport(records, len(records) > 0, chain.tripper))
	return chain
}

func (chain *ChainTransport) WithSRV(domain string, ttl time.Duration) *ChainTransport {
	chain.addConnect(ProviderSRV, NewSRVTransport(domain, ttl, chain.tripper))
	return chain
}

func (chain *ChainTransport) WithSidecar(addr string) *ChainTransport {
	chain.addConnect(ProviderSidecar, NewSidecarTransport(addr, len(addr) > 0, chain.tripper))
	return chain
}

func (chain *ChainTransport) WithConsul(addr string) *ChainTransport {
	chain.addConnect(ProviderConsul, NewConsulTransport(addr, len(addr) > 0, chain.tripper))
	return chain
}

func (chain *ChainTransport) WithRegistry(registry *discovery.Registry) *ChainTransport {
	consulResolver := resolver.NewRegistryResolver(registry, nil)
	chain.addConnect(ProviderConsul, NewTransportWithRegistry(consulResolver, consulResolver != nil, chain.tripper))
	return chain
}

// ReloadConfig merges services defined in both local file and remote config.
//
// NOTE:
// 	- remote config will overwrite local if exists
func (chain *ChainTransport) ReloadConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}

	// for disable case
	if cfg.Disable {
		chain.value.Store(nil)
		return
	}

	var (
		value = chain.load()
	)

	// apply connectors
	allProviders := map[ProviderType]bool{}
	sort.Slice(cfg.Connects, func(i, j int) bool {
		return cfg.Connects[i].Priority < cfg.Connects[j].Priority
	})

	for _, connect := range cfg.Connects {
		if connector := chain.newConnector(string(connect.Provider), connect); connector != nil {
			value.connects.Store(connect.Provider, connector)
		}
	}

	// clean all deprecated connectors
	value.connects.Range(func(key, val interface{}) bool {
		provider, ok := key.(ProviderType)
		if ok && !allProviders[provider] {
			value.connects.Delete(key)
		}

		return true
	})

	// apply services
	allServices := map[string]bool{}
	service2domains := map[string][]string{}
	for _, service := range cfg.Services {
		allServices[service.Name] = true
		value.services.Store(service.Name, service)

		if len(service.Domains) > 0 {
			service2domains[service.Name] = append(service2domains[service.Name], service.Domains...)
		}

		if connector := chain.newConnector(service.Name, service.Connect); connector != nil {
			value.subconnects.Store(service.Name, connector)
		}
	}

	// clean all deprecated services
	value.services.Range(func(key, val interface{}) bool {
		name, ok := key.(string)
		if ok && !allServices[name] {
			value.services.Delete(key)
		}

		return true
	})
	value.subconnects.Range(func(key, val interface{}) bool {
		name, ok := key.(string)
		if ok && !allServices[name] {
			value.subconnects.Delete(key)
		}

		return true
	})

	// build static connect from services
	if len(service2domains) > 0 {
		value.connects.Store(ProviderStatic, NewStaticTransport(service2domains, len(service2domains) > 0, chain.tripper))
	}

	chain.value.Store(value)
}

func (chain *ChainTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// is there a service mapped?
	ctx := req.Context()

	if trace.IsHttpTracingEnabled(req) {
		ctxTrace := trace.New()
		defer func() {
			ctxTrace.Finish()

			chain.logger.Warnf("HTTP Trace %s(%s): %s", req.Method, req.URL.String(), ctxTrace.TraceInfo().String())
		}()

		ctx = ctxTrace.WithContext(ctx)
		*req = *req.WithContext(ctx)
	}

	service, hijacked, err := chain.resolve(ctx, req.URL)
	switch err {
	case nil:
		if service != nil && !hijacked {
			*req = *req.WithContext(ContextWithService(ctx, service))
		}

		tripped := false
		value := chain.load()

		// first, try service connector
		if service.IsConnectValid() {
			iface, ok := value.subconnects.Load(service.Name)
			if ok {
				tripper, ok := iface.(Interface)
				if ok && tripper.IsValid() {
					// nolint: bodyclose
					resp, err = tripper.RoundTrip(req)
					tripped = true
				}
			}
		}

		// second, try global connectors
		if !tripped {
			value.connects.Range(func(key, value interface{}) bool {
				tripper, ok := value.(Interface)
				if ok && tripper.IsValid() {
					// nolint: bodyclose
					resp, err = tripper.RoundTrip(req)
					tripped = true

					return false
				}

				return true
			})
		}

		// finally, fail-back to default tripper if has not tripped or resolver with error
		if !tripped || err == ErrResolverError || err == ErrInvalidTransport {
			// nolint: bodyclose
			resp, err = chain.roundtrip(service, req)
		}

	case ErrServiceNotFound:
		// nolint: bodyclose
		resp, err = chain.roundtrip(service, req)
	}

	// does transform 4xx to http error?
	if err == nil && chain.transform4xx {
		switch resp.StatusCode % 100 {
		case 4:
			buf := bytes.NewBuffer(nil)

			_, ioerr := io.Copy(buf, resp.Body)

			// always close response body
			resp.Body.Close()

			if ioerr != nil {
				err = ioerr
			} else {
				resp.Body = ioutil.NopCloser(buf)

				err = fmt.Errorf("invalid response: code=%d, message=%s", resp.StatusCode, buf.String())
			}
		}
	}

	return
}

func (chain *ChainTransport) Close() {
	if chain.looping == nil {
		return
	}

	close(chain.looping)
}

func (chain *ChainTransport) resolve(ctx context.Context, urlobj *url.URL) (service *config.ServiceConfig, hijacked bool, err error) {
	value := chain.load()

	// first, try service
	service = ContextService(ctx)
	if service != nil {
		hijacked = true
		return
	}

	// second, try service domains
	value.services.Range(func(_, v interface{}) bool {
		svc, ok := v.(*config.ServiceConfig)
		if ok && svc.Match(urlobj) {
			service = svc
			return false
		}

		return true
	})
	if service != nil {
		return
	}

	// third, try service name
	serviceName := ContextServiceName(ctx)
	if len(serviceName) > 0 {
		iface, ok := value.services.Load(serviceName)
		if ok {
			service, ok = iface.(*config.ServiceConfig)
			if ok {
				return
			}
		}

		service = &config.ServiceConfig{
			Name: serviceName,
		}
		return
	}

	err = ErrServiceNotFound
	return
}

func (chain *ChainTransport) roundtrip(service *config.ServiceConfig, req *http.Request) (resp *http.Response, err error) {
	labels := prometheus.Labels{"client": "http", "to": req.URL.Host}
	if service != nil && len(service.Name) > 0 {
		labels["to"] = service.Name
	}

	issuedAt := time.Now()
	resp, err = chain.tripper.RoundTrip(req)
	metrics.ObserveRequest(labels, issuedAt)

	// for http metrics
	metrics.IncRequestTotals(labels)
	if err != nil {
		metrics.IncRequestFailures(labels)

		switch terr := err.(type) {
		case *net.OpError:
			metrics.IncNetFailures(prometheus.Labels{
				"op": terr.Op,
			})

			chain.logger.Errorf("%s(%s): %+v", req.Method, req.URL.String(), terr)
		}
	}

	return
}

func (chain *ChainTransport) loop() {
	chain.looping = make(chan struct{})

	ticker := time.NewTicker(DefaultTTLInterval)
	for {
		select {
		case <-ticker.C:
			value := chain.load()
			value.HealthCheck()

		case <-chain.looping:
			chain.looping = nil
			return
		}
	}
}

func (chain *ChainTransport) addConnect(provider ProviderType, c Interface) {
	value := chain.load()
	value.connects.Store(provider, c)

	chain.value.Store(value)
}

func (chain *ChainTransport) load() *configValue {
	value := chain.value.Load()
	if value == nil {
		return &configValue{}
	}

	cfg, ok := value.(*configValue)
	if !ok {
		return &configValue{}
	}

	return cfg
}

func (chain *ChainTransport) newConnector(service string, connect *config.ConnectConfig) Interface {
	if !connect.IsValid() {
		return nil
	}

	switch connect.Provider {
	case config.ProviderStatic:
		records := map[string][]string{
			service: strings.Split(connect.Addr, ","),
		}

		return NewStaticTransport(records, connect.IsEnabled(), chain.tripper)

	case config.ProviderSidecar:
		// for eds resolver
		if len(connect.EDSAddr) > 0 {
			return NewSidecarTransportWithEDS(connect.Addr, connect.EDSAddr, connect.IsEnabled(), chain.tripper)
		}

		return NewSidecarTransport(connect.Addr, connect.IsEnabled(), chain.tripper)

	case config.ProviderConsul:
		// for eds resolver
		if len(connect.EDSAddr) > 0 {
			return NewConsulTransportWithEDS(connect.Addr, connect.EDSAddr, connect.IsEnabled(), chain.tripper)
		}

		return NewConsulTransport(connect.Addr, connect.IsEnabled(), chain.tripper)
	}

	return nil
}
