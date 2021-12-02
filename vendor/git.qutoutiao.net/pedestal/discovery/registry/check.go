package registry

import "time"

const (
	HealthTypeHTTP = "HTTP"
	HealthTypeTCP  = "TCP"
)

const (
	HealthHTTPServiceNameKey = "X-Qtt-Meshservice"
	HealthHTTPServiceIPKey   = "X-Qtt-Meship"
)

type HealthStatus string

type HealthCheck struct {
	Type string
	Name string
	//HTTP为URI，TCP为地址
	URI    string
	Method string
	//多久检查一次
	Interval time.Duration
	//check初始状态
	Status HealthStatus
	//HTTP 支持header
	Header map[string][]string
}

type TCPHealthCheck struct {
	Name     string
	Addr     string
	Interval time.Duration
	Status   HealthStatus
}

type HTTPHealthCheck struct {
	Name     string
	URI      string
	Interval time.Duration
	Status   HealthStatus
	Method   string
	Header   map[string][]string
}
