package transport

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/golib/resty/retry"

	"git.qutoutiao.net/golib/resty/logger"
	"git.qutoutiao.net/golib/resty/metrics"
	"git.qutoutiao.net/golib/resty/resolver"
	"github.com/prometheus/client_golang/prometheus"
)

type RegistryTransport struct {
	name            string
	tripper         http.RoundTripper
	resolver        resolver.Interface
	logger          logger.Interface
	registryAddr    net.Addr
	registryValue   *atomic.Value
	ttlTotals       uint32
	ttlFailures     uint32
	ttlErrors       uint32
	requestTotals   uint32
	requestFailures uint32
	checkOnce       sync.Once
}

func NewTransportWithRegistry(registry resolver.Interface, enabled bool, tripper http.RoundTripper) *RegistryTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	registryValue := new(atomic.Value)
	registryValue.Store(enabled)
	if registry == nil {
		registryValue.Store(false)
	}

	registryType := fmt.Sprintf("%T", registry)

	rt := &RegistryTransport{
		name:          registryType,
		tripper:       tripper,
		resolver:      registry,
		logger:        logger.NewWithPrefix(registryType),
		registryValue: registryValue,
	}

	return rt
}

func NewConsulTransport(addr string, enabled bool, tripper http.RoundTripper) *RegistryTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	consulResolver := resolver.NewConsulResolver(addr)
	consulLogger := logger.NewWithPrefix("Consul")

	consulValue := new(atomic.Value)
	consulValue.Store(enabled)

	urlobj, err := url.Parse(addr)
	if err != nil {
		consulLogger.Errorf("url.Parse(%s): %+v", addr, err)

		consulValue.Store(false)
	} else {
		if len(urlobj.Host) > 0 {
			addr = urlobj.Host
		}
	}

	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp", addr)
	if tcpErr != nil {
		consulLogger.Errorf("net.ResolveTCPAddr(%s): %+v", addr, tcpErr)

		consulValue.Store(false)
	}

	consul := &RegistryTransport{
		name:          "consul",
		tripper:       tripper,
		resolver:      consulResolver,
		logger:        consulLogger,
		registryAddr:  tcpAddr,
		registryValue: consulValue,
	}
	consul.Check()

	return consul
}

func NewConsulTransportWithEDS(consulAddr, edsAddr string, enabled bool, tripper http.RoundTripper) *RegistryTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	edsResolver := resolver.NewEDSResolver(consulAddr, edsAddr)
	edsLogger := logger.NewWithPrefix("EDS")

	edsValue := new(atomic.Value)
	edsValue.Store(enabled)

	urlobj, err := url.Parse(edsAddr)
	if err != nil {
		edsLogger.Errorf("url.Parse(%s): %+v", edsAddr, err)

		edsValue.Store(false)
	} else {
		if len(urlobj.Host) > 0 {
			edsAddr = urlobj.Host
		}
	}

	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp", edsAddr)
	if tcpErr != nil {
		edsLogger.Errorf("net.ResolveTCPAddr(%s): %+v", edsAddr, err)

		edsValue.Store(false)
	}

	eds := &RegistryTransport{
		name:          "eds",
		tripper:       tripper,
		resolver:      edsResolver,
		logger:        edsLogger,
		registryAddr:  tcpAddr,
		registryValue: edsValue,
	}

	return eds
}

func (registry *RegistryTransport) IsValid() bool {
	// trigger on fly for first time
	registry.checkOnce.Do(func() {
		registry.Check()
	})

	enabled, ok := registry.registryValue.Load().(bool)
	if ok && !enabled {
		return false
	}

	// for ttl stats, mark failure after 10 retries and failures >= totals
	ttlTotals := atomic.LoadUint32(&registry.ttlTotals)
	if ttlTotals <= MinTTLRetries {
		return true
	}

	if atomic.LoadUint32(&registry.ttlFailures) >= ttlTotals {
		return false
	}

	// for request stats, mark failure after 100 retries and failures/totals > 80%
	requestTotals := atomic.LoadUint32(&registry.requestTotals)
	if requestTotals < MinRequestRetries {
		return true
	}

	if atomic.LoadUint32(&registry.requestFailures)*100/requestTotals > MaxRequestFailuresPercent {
		return false
	}

	return true
}

func (registry *RegistryTransport) Check() {
	if registry.registryAddr != nil {
		conn, err := net.DialTimeout("tcp", registry.registryAddr.String(), time.Second)
		if err != nil {
			registry.logger.Errorf("net.DialTimeout(tcp, %s, 1s): %+v", registry.registryAddr.String(), err)
		}

		switch err.(type) {
		case *net.OpError:
			atomic.AddUint32(&registry.ttlTotals, 1)
			atomic.AddUint32(&registry.ttlFailures, 1)
			atomic.AddUint32(&registry.ttlErrors, 1)

			if atomic.LoadUint32(&registry.ttlErrors) >= atomic.LoadUint32(&registry.ttlTotals) {
				registry.registryValue.Store(false)
			}

		default:
			if conn != nil {
				conn.Close()
			}

			// reset ttl stats
			atomic.StoreUint32(&registry.ttlTotals, 0)
			atomic.StoreUint32(&registry.ttlFailures, 0)
			atomic.StoreUint32(&registry.ttlErrors, 0)

			registry.registryValue.Store(true)
		}
	} else {
		// reset ttl stats
		atomic.StoreUint32(&registry.ttlTotals, 0)
		atomic.StoreUint32(&registry.ttlFailures, 0)
		atomic.StoreUint32(&registry.ttlErrors, 0)

		registry.registryValue.Store(true)
	}

	return
}

func (registry *RegistryTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if !registry.IsValid() {
		err = ErrInvalidTransport
		return
	}

	ctx := req.Context()
	if err = ctx.Err(); err != nil {
		return
	}

	ctxService := ContextService(ctx)
	if ctxService == nil {
		err = ErrServiceNotFound
		return
	}

	labels := prometheus.Labels{"client": registry.name, "to": ctxService.Name}

	var (
		retryMax    = 3
		retryPeriod = 3 * time.Millisecond
		retryFactor = 1.5
	)

	retry.JitterRetry(func(retries uint32) error {
		atomic.AddUint32(&registry.ttlTotals, 1)
		atomic.AddUint32(&registry.requestTotals, 1)

		// metrics for resolver
		metrics.IncResolverTotals(labels)

		service, tmperr := registry.resolver.Resolve(ctx, ctxService.Name)
		if tmperr != nil {
			registry.logger.Errorf("%T.Resolve(%+v): %+v", registry.resolver, ctxService, tmperr)

			// metrics for resolver error
			metrics.IncResolverFailures(labels)

			switch tmperr.(type) {
			case *net.OpError:
				atomic.AddUint32(&registry.ttlFailures, 1)
				atomic.AddUint32(&registry.ttlErrors, 1)
			}

			// for chain transport fallback case
			err = ErrResolverError

			// no retry for resolver error
			return nil
		}

		// metrics for request
		metrics.IncRequestTotals(labels)

		// adjust url host requested
		addr := ctxService.NormalizeAddr(service.Addr())
		switch req.URL.Scheme {
		case "https":
			req.URL.Scheme = "http"
			req.URL.Host = addr
		default:
			req.URL.Host = addr
		}

		issuedAt := time.Now()
		// nolint: bodyclose
		resp, err = registry.tripper.RoundTrip(req)
		metrics.ObserveRequest(labels, issuedAt) // metrics for request latency

		if err == nil {
			return nil
		}

		// metrics for response error
		metrics.IncRequestFailures(labels)
		atomic.AddUint32(&registry.requestFailures, 1)

		// NOTE: only retry with dial error
		switch terr := err.(type) {
		case *net.OpError:
			registry.logger.Errorf("%s(%s): service=%+v, error=%+v", req.Method, req.URL.String(), ctxService, terr)

			// metrics for net.OpError
			metrics.IncNetFailures(prometheus.Labels{
				"op": terr.Op,
			})

			if terr.Op == "dial" {
				registry.resolver.Block(ctx, ctxService.Name, service)

				return err
			}
		}

		return nil
	}, retryMax, retryPeriod, retryFactor)

	if err != nil {
		return
	}

	// mark failures
	switch resp.StatusCode {
	case http.StatusGatewayTimeout,
		http.StatusGone,
		http.StatusBadGateway,
		http.StatusHTTPVersionNotSupported,
		http.StatusLoopDetected,
		499:
		registry.logger.Errorf("%s(%s): service=%+v, status=%d", req.Method, req.URL.String(), ctxService, resp.StatusCode)

		atomic.AddUint32(&registry.requestFailures, 1)

		// metrics for response failures
		metrics.IncRequestFailures(labels)
	}

	return
}
