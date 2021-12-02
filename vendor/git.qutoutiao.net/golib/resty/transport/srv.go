package transport

import (
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/golib/resty/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"git.qutoutiao.net/golib/resty/logger"
	"git.qutoutiao.net/golib/resty/resolver"
)

type SRVTransport struct {
	tripper         http.RoundTripper
	resolver        resolver.Interface
	logger          logger.Interface
	srvAddr         string
	srvValue        *atomic.Value
	ttlTotals       uint32
	ttlFailures     uint32
	ttlRetries      uint32
	requestTotals   uint32
	requestFailures uint32
}

func NewSRVTransport(domain string, ttl time.Duration, tripper http.RoundTripper) *SRVTransport {
	srvResolver := resolver.NewSRVResolver(domain, ttl)

	if tripper == nil {
		tripper = New(nil)
	}

	srv := &SRVTransport{
		tripper:  tripper,
		resolver: srvResolver,
		logger:   logger.NewWithPrefix("SRV"),
		srvAddr:  domain,
		srvValue: new(atomic.Value),
	}
	srv.Check()

	return srv
}

func (srv *SRVTransport) IsValid() bool {
	enabled, ok := srv.srvValue.Load().(bool)
	if ok && !enabled {
		return false
	}

	// for ttl stats, mark failure after 10 retries and failures >= totals
	ttlTotals := atomic.LoadUint32(&srv.ttlTotals)
	if ttlTotals < MinTTLRetries {
		return true
	}

	if atomic.LoadUint32(&srv.ttlFailures) >= ttlTotals {
		return false
	}

	// for request stats, mark failure after 100 retries and failures/totals > 80%
	requestTotals := atomic.LoadUint32(&srv.requestTotals)
	if requestTotals < MinRequestRetries {
		return true
	}

	if atomic.LoadUint32(&srv.requestFailures)*100/requestTotals > MaxRequestFailuresPercent {
		return false
	}

	return true
}

// TODO: check with SRV resolver?
func (srv *SRVTransport) Check() {
	if atomic.LoadUint32(&srv.ttlFailures) == 0 || atomic.LoadUint32(&srv.ttlFailures) < atomic.LoadUint32(&srv.ttlTotals) {
		srv.srvValue.Store(true)
	} else {
		srv.srvValue.Store(false)
	}

	// reset ttl stats
	if atomic.AddUint32(&srv.ttlRetries, 1) > MinTTLRetries {
		atomic.StoreUint32(&srv.ttlTotals, 0)
		atomic.StoreUint32(&srv.ttlFailures, 0)
		atomic.StoreUint32(&srv.ttlRetries, 0)
	}

	return
}

func (srv *SRVTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if !srv.IsValid() {
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

	atomic.AddUint32(&srv.ttlTotals, 1)
	atomic.AddUint32(&srv.requestTotals, 1)

	// for metrics
	labels := prometheus.Labels{"client": "SRV", "to": ctxService.Name}

	metrics.IncResolverTotals(labels)
	service, err := srv.resolver.Resolve(ctx, ctxService.Name)
	if err != nil {
		srv.logger.Errorf("resolve service(%s): %v", ctxService, err)

		// error metrics of resolver
		metrics.IncResolverFailures(labels)

		switch err.(type) {
		case *net.OpError:
			atomic.AddUint32(&srv.ttlFailures, 1)
		}

		// for fallback issue
		err = ErrResolverError
		return
	}

	// adjust url host requested
	addr := service.Addr()
	switch req.URL.Scheme {
	case "https":
		req.URL.Scheme = "http"
		req.URL.Host = addr
	default:
		req.URL.Host = addr
	}

	// NOTE: SRV transport is not connect!
	metrics.IncRequestTotals(labels)

	issuedAt := time.Now()
	resp, err = srv.tripper.RoundTrip(req)
	metrics.ObserveRequest(labels, issuedAt)

	if err != nil {
		atomic.AddUint32(&srv.requestFailures, 1)

		// error metrics of request
		metrics.IncRequestFailures(labels)

		switch terr := err.(type) {
		case *net.OpError:
			metrics.IncNetFailures(prometheus.Labels{
				"op": terr.Op,
			})

			srv.logger.Errorf("connect(%s): %+v", req.URL.String(), terr)
		}
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
		srv.logger.Errorf("%s(%s): service=%s, status=%d", req.Method, req.URL.String(), ctxService, resp.StatusCode)

		atomic.AddUint32(&srv.requestFailures, 1)

		// error metrics of request
		metrics.IncRequestFailures(labels)
	}
	return
}
