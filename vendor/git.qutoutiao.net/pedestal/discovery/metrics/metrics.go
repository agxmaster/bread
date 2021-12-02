package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	prom *Metrics
	once sync.Once
)

type Metrics struct {
	consulDegrade   *prometheus.GaugeVec
	consulThreshold *prometheus.GaugeVec
	totalNodes      *prometheus.GaugeVec
	client          *prometheus.CounterVec
	latency         *prometheus.HistogramVec
	version         *prometheus.GaugeVec
}

func GetMetrics() *Metrics {
	once.Do(func() {
		prom = NewMetrics()
		prom.Register(prometheus.Register)
	})

	return prom
}

func NewMetrics() *Metrics {
	// 报告服务是否处于降级状态
	degrade := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "discovery",
		Subsystem: "consul",
		Name:      "degrade",
		Help:      "Report degrade status of service with consul discovery",
	}, []string{"service_name"})

	threshold := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "discovery",
		Subsystem: "consul",
		Name:      "threshold",
		Help:      "Report degrade threshold of service with consul discovery",
	}, []string{"service_name"})

	nodes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "discovery",
		Subsystem: "total",
		Name:      "nodes",
		Help:      "Report total nodes of service resolved",
	}, []string{"adapter", "service_name"})

	client := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "discovery",
		Subsystem: "client",
		Name:      "request",
		Help:      "Report total request of service with discovery client",
	}, []string{"service_name", "client_type"})

	latency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "discovery",
		Subsystem: "client",
		Name:      "latency",
		Help:      "Report latency of service with discovery client in seconds",
	}, []string{"service_name", "client_type"})

	version := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "discovery",
		Subsystem: "client",
		Name:      "version",
		Help:      "Report sdk version of service imported",
	}, []string{"client_version"})

	return &Metrics{
		consulDegrade:   degrade,
		consulThreshold: threshold,
		totalNodes:      nodes,
		client:          client,
		latency:         latency,
		version:         version,
	}
}

// register all metrics
func (metrics *Metrics) Register(fn func(prometheus.Collector) error) {
	fn(metrics.consulDegrade)
	fn(metrics.consulThreshold)
	fn(metrics.totalNodes)
	fn(metrics.client)
	fn(metrics.latency)
	fn(metrics.version)
}

func (metrics *Metrics) ReportClientVersion(version string) {
	metrics.version.WithLabelValues(version).Set(1)
}

func (metrics *Metrics) ReportConsulDegrade(service string, isDegrade bool) {
	if isDegrade {
		metrics.consulDegrade.WithLabelValues(service).Set(1)
		return
	}

	metrics.consulDegrade.WithLabelValues(service).Set(0)
}

func (metrics *Metrics) ReportConsulThreshold(service string, num int) {
	metrics.consulThreshold.WithLabelValues(service).Set(float64(num))
}

func (metrics *Metrics) ReportTotalNodes(adapter, service string, num int) {
	metrics.totalNodes.WithLabelValues(adapter, service).Set(float64(num))
}

func (metrics *Metrics) IncRequestCounter(service string, clientType string) {
	metrics.client.WithLabelValues(service, clientType).Inc()
}

func (metrics *Metrics) ObserveRequestLatency(service string, clientType string, issuedAt time.Time) {
	metrics.latency.WithLabelValues(service, clientType).Observe(time.Since(issuedAt).Seconds())
}
