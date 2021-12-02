package metrics

import (
	"sync"

	"git.qutoutiao.net/golib/resty/config"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	restyVersionGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "resty",
		Subsystem: "client",
		Name:      "version",
		Help:      "Total success of resolver",
	}, []string{"version"})
	resolverTotalsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "resty",
		Subsystem: "resolver",
		Name:      "totals",
		Help:      "Total success of resolver",
	}, []string{"client", "to"})
	resolverFailuresCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "resty",
		Subsystem: "resolver",
		Name:      "failures",
		Help:      "Total failures of resolver",
	}, []string{"client", "to"})
	netFailuresCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "resty",
		Subsystem: "net",
		Name:      "failures",
		Help:      "Total failures of net.OpError",
	}, []string{"op"})
	requestTotalsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "resty",
		Subsystem: "request",
		Name:      "totals",
		Help:      "Total success of request",
	}, []string{"client", "to"})
	requestFailuresCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "resty",
		Subsystem: "request",
		Name:      "failures",
		Help:      "Total failures of request",
	}, []string{"client", "to"})
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "resty",
		Subsystem: "request",
		Name:      "latency",
		Help:      "avg latency of request, unit in seconds",
	}, []string{"client", "to"})

	registerOnce sync.Once
)

func init() {
	registerOnce.Do(func() {
		restyVersionGauge.With(prometheus.Labels{
			"version": config.Version,
		}).Set(1)

		prometheus.MustRegister(restyVersionGauge)
		prometheus.MustRegister(resolverTotalsCounter)
		prometheus.MustRegister(resolverFailuresCounter)
		prometheus.MustRegister(netFailuresCounter)
		prometheus.MustRegister(requestTotalsCounter)
		prometheus.MustRegister(requestFailuresCounter)
		prometheus.MustRegister(requestLatency)
	})
}
