package trace

import (
	"encoding/json"
	"fmt"
	"net/http/httptrace"
	"time"
)

// HttpTraceInfo struct is used provide request enableHttpTrace info such as DNS lookup
// duration, Connection obtain duration, Server processing duration, etc.
type HttpTraceInfo struct {
	// Latency is total duration that request took end-to-end.
	Latency time.Duration

	ConnInfo httptrace.GotConnInfo
	DNSInfo  httptrace.DNSDoneInfo
	PeerInfo string

	// DNSLookup is a duration that transport took to perform
	// DNS lookup.
	DNSLookup time.Duration

	// ConnectTime is a duration that took to obtain a successful connection.
	ConnectTime  time.Duration
	ConnectError error

	// TLSHandshake is a duration that TLS handshake took place.
	TLSHandshake time.Duration

	// RequestTime is a duration that server took to respond first byte.
	RequestTime  time.Duration
	RequestError error

	// ResponseTime is a duration since first response byte from server to
	// request completion.
	ResponseTime  time.Duration
	ResponseError error
}

func (info HttpTraceInfo) String() string {
	m := map[string]string{
		"Latency":       fmt.Sprintf("%v", info.Latency),
		"DNSInfo":       fmt.Sprintf("%+v", info.DNSInfo),
		"ConnInfo":      fmt.Sprintf("%+v", info.ConnInfo),
		"PeerInfo":      info.PeerInfo,
		"DNSLookup":     fmt.Sprintf("%v", info.DNSLookup),
		"ConnectTime":   fmt.Sprintf("%v", info.ConnectTime),
		"ConnectError":  fmt.Sprintf("%+v", info.ConnectError),
		"TLSHandshake":  fmt.Sprintf("%v", info.TLSHandshake),
		"RequestTime":   fmt.Sprintf("%v", info.RequestTime),
		"RequestError":  fmt.Sprintf("%+v", info.RequestError),
		"ResponseTime":  fmt.Sprintf("%v", info.ResponseTime),
		"ResponseError": fmt.Sprintf("%+v", info.ResponseError),
	}

	b, _ := json.Marshal(m)
	return string(b)
}
