package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	gormVersionGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "gorm",
		Subsystem: "sdk",
		Name:      "version",
		Help:      "Current gorm version",
	}, []string{"version"})
	gormCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gorm",
		Name:      "count",
		Help:      "Total number of request gorm.",
	}, []string{"client", "cmd", "to", "status"})
	gormLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gorm",
		Name:      "duration",
		Help:      "avg latency of operations, unit in seconds",
	}, []string{"client", "cmd", "to", "status"})

	registerOnce sync.Once
)

func init() {
	registerOnce.Do(func() {
		gormVersionGauge.With(prometheus.Labels{
			"version": Version,
		}).Set(1)

		prometheus.MustRegister(gormVersionGauge)
		prometheus.MustRegister(gormCounter)
		prometheus.MustRegister(gormLatency)
	})
}
