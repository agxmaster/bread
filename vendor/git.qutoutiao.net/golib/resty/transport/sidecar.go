package transport

import (
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/golib/resty/logger"
	"git.qutoutiao.net/golib/resty/metrics"
	"git.qutoutiao.net/golib/resty/resolver"
	"git.qutoutiao.net/golib/resty/retry"
	"github.com/prometheus/client_golang/prometheus"
)

type SidecarTransport struct {
	tripper         http.RoundTripper
	resolver        resolver.Interface
	logger          logger.Interface
	sidecarAddr     net.Addr
	sidecarValue    *atomic.Value
	ttlTotals       uint32
	ttlFailures     uint32
	ttlErrors       uint32
	requestTotals   uint32
	requestFailures uint32
	checkOnce       sync.Once
}

func NewSidecarTransport(addr string, enabled bool, tripper http.RoundTripper) *SidecarTransport {
	return NewSidecarTransportWithEDS(addr, "", enabled, tripper)
}

func NewSidecarTransportWithEDS(sidecarAddr, edsAddr string, enabled bool, tripper http.RoundTripper) *SidecarTransport {
	if tripper == nil {
		tripper = New(nil)
	}

	sidecarLogger := logger.NewWithPrefix("Sidecar")

	sidecarValue := new(atomic.Value)
	sidecarValue.Store(enabled)

	urlobj, err := url.Parse(sidecarAddr)
	if err != nil {
		sidecarLogger.Errorf("url.Parse(%s): %+v", sidecarAddr, err)

		sidecarValue.Store(false)
	} else {
		if len(urlobj.Host) > 0 {
			sidecarAddr = urlobj.Host
		}
	}

	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp", sidecarAddr)
	if tcpErr != nil {
		sidecarLogger.Errorf("net.ResolveTCPAddr(%s): %+v", sidecarAddr, tcpErr)

		sidecarValue.Store(false)
	}

	sidecar := &SidecarTransport{
		tripper:      tripper,
		resolver:     resolver.NewSidecarResolverWithEDS(sidecarAddr, edsAddr),
		logger:       sidecarLogger,
		sidecarAddr:  tcpAddr,
		sidecarValue: sidecarValue,
	}
	sidecar.Check()

	return sidecar
}

func (sidecar *SidecarTransport) IsValid() bool {
	// trigger on fly for first time
	sidecar.checkOnce.Do(func() {
		sidecar.Check()
	})

	enabled, ok := sidecar.sidecarValue.Load().(bool)
	if ok && !enabled {
		return false
	}

	// for ttl stats, mark failure after 10 retries and failures >= totals
	ttlTotals := atomic.LoadUint32(&sidecar.ttlTotals)
	if ttlTotals <= MinTTLRetries {
		return true
	}

	if atomic.LoadUint32(&sidecar.ttlFailures) >= ttlTotals {
		return false
	}

	// for request stats, mark failure after 100 retries and failures/totals > 80%
	requestTotals := atomic.LoadUint32(&sidecar.requestTotals)
	if requestTotals < MinRequestRetries {
		return true
	}

	if atomic.LoadUint32(&sidecar.requestFailures)*100/requestTotals > MaxRequestFailuresPercent {
		return false
	}

	return true
}

func (sidecar *SidecarTransport) Check() {
	if sidecar.sidecarAddr != nil {
		conn, err := net.DialTimeout("tcp", sidecar.sidecarAddr.String(), time.Second)
		if err != nil {
			sidecar.logger.Errorf("net.DialTimeout(tcp, %s, 1s): %+v", sidecar.sidecarAddr.String(), err)
		}

		switch err.(type) {
		case *net.OpError:
			atomic.AddUint32(&sidecar.ttlTotals, 1)
			atomic.AddUint32(&sidecar.ttlFailures, 1)
			atomic.AddUint32(&sidecar.ttlErrors, 1)

			if atomic.LoadUint32(&sidecar.ttlErrors) >= atomic.LoadUint32(&sidecar.ttlTotals) {
				sidecar.sidecarValue.Store(false)
			}

		default:
			if conn != nil {
				conn.Close()
			}

			// reset ttl stats
			atomic.StoreUint32(&sidecar.ttlTotals, 0)
			atomic.StoreUint32(&sidecar.ttlFailures, 0)
			atomic.StoreUint32(&sidecar.ttlErrors, 0)

			sidecar.sidecarValue.Store(true)
		}
	} else {
		// reset ttl stats
		atomic.StoreUint32(&sidecar.ttlTotals, 0)
		atomic.StoreUint32(&sidecar.ttlFailures, 0)
		atomic.StoreUint32(&sidecar.ttlErrors, 0)

		sidecar.sidecarValue.Store(true)
	}

	return
}

func (sidecar *SidecarTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if !sidecar.IsValid() {
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

	labels := prometheus.Labels{"client": "sidecar", "to": ctxService.Name}

	atomic.AddUint32(&sidecar.ttlTotals, 1)
	atomic.AddUint32(&sidecar.requestTotals, 1)

	var (
		retryMax    = 3
		retryPeriod = 3 * time.Millisecond
		retryFactor = 1.5
	)

	retry.JitterRetry(func(retries uint32) error {
		atomic.AddUint32(&sidecar.ttlTotals, 1)
		atomic.AddUint32(&sidecar.requestTotals, 1)

		// metrics for resolver
		metrics.IncResolverTotals(labels)

		service, tmperr := sidecar.resolver.Resolve(ctx, ctxService.Name)
		if tmperr != nil {
			sidecar.logger.Errorf("%T.Resolve(%+v): %+v", sidecar.resolver, ctxService, tmperr)

			// metrics for resolver error
			metrics.IncResolverFailures(labels)

			switch tmperr.(type) {
			case *net.OpError:
				atomic.AddUint32(&sidecar.ttlFailures, 1)
				atomic.AddUint32(&sidecar.ttlErrors, 1)
			}

			// for chain transport fallback case
			err = ErrResolverError

			// no retry for resolver error
			return nil
		}

		// metrics for request
		metrics.IncRequestTotals(labels)

		// adjust url host requested
		addr := service.Addr()
		switch req.URL.Scheme {
		case "https":
			req.URL.Scheme = "http"
			req.URL.Host = addr
		default:
			req.URL.Host = addr
		}

		// for sidecar connect
		switch service.ID {
		case resolver.DefaultSidecarServiceID:
			req.Host = addr
			req.RequestURI = req.URL.String()

			// apply service resolver meta
			req.Header.Set(resolver.SidecarHeaderDatacenterKey, ctxService.DC)
			req.Header.Set(resolver.SidecarHeaderServiceKey, ctxService.Name)
			for _, tag := range ctxService.Tags {
				req.Header.Add(resolver.SidecarHeaderTagsKey, tag)
			}
		}

		issuedAt := time.Now()
		// nolint: bodyclose
		resp, err = sidecar.tripper.RoundTrip(req)
		metrics.ObserveRequest(labels, issuedAt) // metrics for request latency

		if err == nil {
			return nil
		}

		// metrics for response error
		metrics.IncRequestFailures(labels)
		atomic.AddUint32(&sidecar.requestFailures, 1)

		// NOTE: only retry with dial error
		switch terr := err.(type) {
		case *net.OpError:
			sidecar.logger.Errorf("%s(%s): service=%+v, error=%+v", req.Method, req.URL.String(), ctxService, terr)

			// metrics for net.OpError
			metrics.IncNetFailures(prometheus.Labels{
				"op": terr.Op,
			})

			if terr.Op == "dial" {
				sidecar.resolver.Block(ctx, ctxService.Name, service)

				return err
			}
		}

		return nil
	}, retryMax, retryPeriod, retryFactor)

	if err != nil {
		return
	}

	// for sidecar response errors
	sidecarErr := resp.Header.Get(resolver.SidecarHeaderErrorKey)
	if len(sidecarErr) > 0 {
		sidecar.logger.Errorf("%s(%s): service=%+v, sidecar=%+v", req.Method, req.URL.String(), ctxService, sidecarErr)
	}

	// mark failures
	switch resp.StatusCode {
	case http.StatusGatewayTimeout,
		http.StatusGone,
		http.StatusBadGateway,
		http.StatusHTTPVersionNotSupported,
		http.StatusLoopDetected,
		499:
		sidecar.logger.Errorf("%s(%s): connect=sidecar, service=%+v, status=%d", req.Method, req.URL.String(), ctxService, resp.StatusCode)

		atomic.AddUint32(&sidecar.requestFailures, 1)

		// error metrics of request
		metrics.IncRequestFailures(labels)
	}

	return
}
