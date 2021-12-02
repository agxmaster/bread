package transport

import (
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/golib/resty/logger"
	"git.qutoutiao.net/golib/resty/metrics"
	"git.qutoutiao.net/golib/resty/resolver"
	"git.qutoutiao.net/golib/resty/retry"
	"github.com/prometheus/client_golang/prometheus"
)

type StaticTransport struct {
	tripper         http.RoundTripper
	resolver        resolver.Interface
	logger          logger.Interface
	staticValue     *atomic.Value
	ttlTotals       uint32
	ttlFailures     uint32
	ttlRetries      uint32
	requestTotals   uint32
	requestFailures uint32
}

func NewStaticTransport(records map[string][]string, enabled bool, tripper http.RoundTripper) *StaticTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	staticValue := new(atomic.Value)
	staticValue.Store(enabled)

	static := &StaticTransport{
		tripper:     tripper,
		resolver:    resolver.NewStaticResolver(records),
		logger:      logger.NewWithPrefix("Static"),
		staticValue: staticValue,
	}
	static.Check()

	return static
}

func (static *StaticTransport) IsValid() bool {
	enabled, ok := static.staticValue.Load().(bool)
	if ok && !enabled {
		return false
	}

	// for ttl stats, mark failure after 10 retries
	ttlTotals := atomic.LoadUint32(&static.ttlTotals)
	if ttlTotals < MinTTLRetries {
		return true
	}

	return atomic.LoadUint32(&static.ttlFailures) < ttlTotals
}

// TODO: check with static resolver?
func (static *StaticTransport) Check() {
	if atomic.LoadUint32(&static.ttlFailures) == 0 || atomic.LoadUint32(&static.ttlFailures) < atomic.LoadUint32(&static.ttlTotals) {
		static.staticValue.Store(true)
	} else {
		static.staticValue.Store(false)
	}

	// reset ttl stats
	if atomic.AddUint32(&static.ttlRetries, 1) > MinTTLRetries {
		atomic.StoreUint32(&static.ttlTotals, 0)
		atomic.StoreUint32(&static.ttlFailures, 0)
		atomic.StoreUint32(&static.ttlRetries, 0)
	}

	return
}

func (static *StaticTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if !static.IsValid() {
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

	labels := prometheus.Labels{"client": "static", "to": ctxService.Name}

	var (
		retryMax    = 3
		retryPeriod = 3 * time.Millisecond
		retryFactor = 1.5
	)

	retry.JitterRetry(func(retries uint32) error {
		atomic.AddUint32(&static.ttlTotals, 1)
		atomic.AddUint32(&static.requestTotals, 1)

		// metrics for resolver
		metrics.IncResolverTotals(labels)

		service, tmperr := static.resolver.Resolve(ctx, ctxService.Name)
		if tmperr != nil {
			static.logger.Errorf("%T.Resolve(%+v): %+v", static.resolver, ctxService, tmperr)

			// metrics for resolver error
			metrics.IncResolverFailures(labels)

			atomic.AddUint32(&static.ttlFailures, 1)

			// for chain transport fallback case
			err = ErrResolverError

			// no retry for resolver error
			return nil
		}

		// metrics for request
		metrics.IncRequestTotals(labels)

		// adjust url host requested
		if ctxService.Port > 0 {
			service.Port = ctxService.Port
		}
		req.URL.Host = service.Addr()

		issuedAt := time.Now()
		// nolint: bodyclose
		resp, err = static.tripper.RoundTrip(req)
		metrics.ObserveRequest(labels, issuedAt) // metrics for request latency

		if err == nil {
			return nil
		}

		// metrics for response error
		metrics.IncRequestFailures(labels)
		atomic.AddUint32(&static.requestFailures, 1)

		// NOTE: only retry with dial error
		switch terr := err.(type) {
		case *net.OpError:
			static.logger.Errorf("%s(%s): service=%+v, error=%+v", req.Method, req.URL.String(), ctxService, terr)

			metrics.IncNetFailures(prometheus.Labels{
				"op": terr.Op,
			})

			if terr.Op == "dial" {
				static.resolver.Block(ctx, ctxService.Name, service)

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
		static.logger.Errorf("%s(%s): service=%+v, status=%d", req.Method, req.URL.String(), ctxService, resp.StatusCode)

		atomic.AddUint32(&static.requestFailures, 1)

		// error metrics of request
		metrics.IncRequestFailures(labels)
	}
	return
}
