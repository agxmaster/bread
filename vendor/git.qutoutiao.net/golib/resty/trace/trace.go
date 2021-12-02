package trace

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

const (
	HTTPTraceHeaderKey = "X-Http-Trace"
	HTTPTraceEnvKey    = "X_HTTP_TRACE"
)

func IsHttpTracingEnabled(req *http.Request) bool {
	return req.Header.Get(HTTPTraceHeaderKey) == "enable" || os.Getenv(HTTPTraceEnvKey) == "enable"
}

// Trace struct maps the `httptrace.ClientTrace` hooks into Fields
// with same naming for easy understanding. Plus additional insights
// Request.
type Trace struct {
	getConn              time.Time
	gotConn              time.Time
	gotFirstResponseByte time.Time
	dnsStart             time.Time
	dnsDone              time.Time
	connectStart         time.Time
	connectDone          time.Time
	connectError         error
	tlsHandshakeStart    time.Time
	tlsHandshakeDone     time.Time
	wroteHeaders         time.Time
	wroteRequest         time.Time
	wroteRequestError    error
	dnsInfo              httptrace.DNSDoneInfo
	connInfo             httptrace.GotConnInfo
	peerInfo             string
	issuedAt             time.Time
	finishedAt           time.Time
}

func New() *Trace {
	return &Trace{
		issuedAt: time.Now(),
	}
}

func (trace *Trace) WithContext(ctx context.Context) context.Context {
	return httptrace.WithClientTrace(
		ctx,
		&httptrace.ClientTrace{
			GetConn: func(_ string) {
				trace.getConn = time.Now()
			},
			GotConn: func(ci httptrace.GotConnInfo) {
				trace.gotConn = time.Now()
				trace.connInfo = ci
			},
			GotFirstResponseByte: func() {
				trace.gotFirstResponseByte = time.Now()
			},
			DNSStart: func(_ httptrace.DNSStartInfo) {
				trace.dnsStart = time.Now()
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				trace.dnsDone = time.Now()
				trace.dnsInfo = info
			},
			ConnectStart: func(network, addr string) {
				trace.connectStart = time.Now()
			},
			ConnectDone: func(network, addr string, err error) {
				trace.connectDone = time.Now()
				trace.connectError = err
				trace.peerInfo = fmt.Sprintf("%s(%s)", network, addr)
			},
			TLSHandshakeStart: func() {
				trace.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				trace.tlsHandshakeDone = time.Now()
			},
			WroteHeaders: func() {
				trace.wroteHeaders = time.Now()
			},
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				trace.wroteRequest = time.Now()
				trace.wroteRequestError = info.Err
			},
		},
	)
}

func (trace *Trace) Latency() time.Duration {
	if trace == nil {
		return 0
	}

	return trace.issuedAt.Sub(trace.getConn)
}

func (trace *Trace) DNSInfo() httptrace.DNSDoneInfo {
	return trace.dnsInfo
}

func (trace *Trace) ConnInfo() httptrace.GotConnInfo {
	return trace.connInfo
}

func (trace *Trace) PeerInfo() string {
	return trace.peerInfo
}

func (trace *Trace) ConnectError() error {
	return trace.connectError
}

func (trace *Trace) RequestError() error {
	return trace.wroteRequestError
}

func (trace *Trace) Finish() {
	if trace == nil || !trace.finishedAt.IsZero() {
		return
	}

	trace.finishedAt = time.Now()
}

// TraceInfo returns the HttpTraceInfo for the request.
// If either the Client or Request WithHttpTracing function has not been called
// prior to the request being made, an empty HttpTraceInfo object will be returned.
func (trace *Trace) TraceInfo() HttpTraceInfo {
	if trace == nil {
		return HttpTraceInfo{}
	}

	return HttpTraceInfo{
		Latency:       trace.finishedAt.Sub(trace.getConn),
		DNSInfo:       trace.dnsInfo,
		ConnInfo:      trace.connInfo,
		PeerInfo:      trace.peerInfo,
		DNSLookup:     trace.dnsDone.Sub(trace.dnsStart),
		ConnectTime:   trace.gotConn.Sub(trace.getConn),
		ConnectError:  trace.connectError,
		TLSHandshake:  trace.tlsHandshakeDone.Sub(trace.tlsHandshakeStart),
		RequestTime:   trace.gotFirstResponseByte.Sub(trace.wroteRequest),
		RequestError:  trace.wroteRequestError,
		ResponseTime:  trace.finishedAt.Sub(trace.gotFirstResponseByte),
		ResponseError: nil,
	}
}
