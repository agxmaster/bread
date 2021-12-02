package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	redisVersionGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "redis",
		Subsystem: "sdk",
		Name:      "version",
		Help:      "Current gorm version",
	}, []string{"version"})
	redisConnPoolCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "redis",
		Subsystem: "conn",
		Name:      "pool",
		Help:      "Number of connection was found in the pool.",
	}, []string{"stats", "to"})
	redisCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "redis",
		Name:      "count",
		Help:      "Total number of request redis.",
	}, []string{"cmd", "to", "status"})
	redisLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "redis",
		Name:      "duration",
		Help:      "avg latency of commands, unit in seconds",
	}, []string{"cmd", "to", "status"})

	registerOnce sync.Once
)

func init() {
	registerOnce.Do(func() {
		redisVersionGauge.With(prometheus.Labels{
			"version": Version,
		}).Set(1)

		prometheus.MustRegister(redisVersionGauge)
		prometheus.MustRegister(redisConnPoolCounter)
		prometheus.MustRegister(redisCounter)
		prometheus.MustRegister(redisLatency)
	})
}
