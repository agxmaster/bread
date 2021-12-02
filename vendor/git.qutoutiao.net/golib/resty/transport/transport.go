package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	DefaultDialTimeout     = 5 * time.Second
	DefaultDialKeepalive   = 90 * time.Second
	DefaultMaxIdleConns    = 100
	DefaultIdleConnTimeout = 90 * time.Second
	DefaultTTLInterval     = 6 * time.Second // 10 times per minute
)

const (
	MinTTLRetries             = 10  // counter
	MinRequestRetries         = 100 // counter
	MaxRequestFailuresPercent = 80  // percentage
)

func New(localAddr net.Addr) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   DefaultDialTimeout,
		KeepAlive: DefaultDialKeepalive,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          DefaultMaxIdleConns,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		MaxConnsPerHost:       DefaultMaxIdleConns,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) * 2,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func NewWithTimeout(localAddr net.Addr, responseHeaderTimeout time.Duration) *http.Transport {
	ts := New(localAddr)
	ts.ResponseHeaderTimeout = responseHeaderTimeout
	ts.ExpectContinueTimeout = responseHeaderTimeout

	return ts
}
