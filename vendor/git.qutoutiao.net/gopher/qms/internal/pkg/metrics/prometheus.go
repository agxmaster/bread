package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//PrometheusExporter is a prom exporter for qms
type PrometheusExporter struct {
	FlushInterval time.Duration
	counters      sync.Map
	gauges        sync.Map
	summaries     sync.Map
	histograms    sync.Map
	limit         *labelLimit
}

//NewPrometheusExporter create a prometheus exporter
func NewPrometheusExporter(options Options) Registry {
	return &PrometheusExporter{
		FlushInterval: options.FlushInterval,
		summaries:     sync.Map{},
		counters:      sync.Map{},
		gauges:        sync.Map{},
		histograms:    sync.Map{},
		limit:         newLabelLimit(),
	}
}

//CreateGauge create collector
func (c *PrometheusExporter) CreateGauge(opts *GaugeOpts) error {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)

	if _, ok := c.gauges.Load(opts.Name); ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.gauges.Store(opts.Name, vec)

	prometheus.MustRegister(vec)
	return nil
}

//GaugeSet set value
func (c *PrometheusExporter) GaugeSet(name string, val float64, labels map[string]string) error {
	if !c.limit.safeCheck(name, labels) {
		return fmt.Errorf("labels up to limit")
	}

	iface, ok := c.gauges.Load(name)
	if !ok {
		return fmt.Errorf("metrics [%s] do not exists, create it first", name)
	}

	vec, ok := iface.(*prometheus.GaugeVec)
	if !ok {
		c.gauges.Delete(name)

		return fmt.Errorf("metrics [%s] is invalid, create it first", name)
	}

	vec.With(labels).Set(val)
	return nil
}

//GaugeReset reset value
func (c *PrometheusExporter) GaugeReset(name string) error {
	iface, ok := c.gauges.Load(name)
	if !ok {
		return fmt.Errorf("metrics [%s] do not exists, create it first", name)
	}

	vec, ok := iface.(*prometheus.GaugeVec)
	if !ok {
		c.gauges.Delete(name)

		return fmt.Errorf("metrics [%s] is invalid, create it first", name)
	}

	vec.Reset()
	return nil
}

//CreateCounter create collector
func (c *PrometheusExporter) CreateCounter(opts *CounterOpts) error {
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)

	if _, ok := c.counters.Load(opts.Name); ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.counters.Store(opts.Name, vec)

	prometheus.MustRegister(vec)
	return nil
}

//CounterAdd increase value
func (c *PrometheusExporter) CounterAdd(name string, val float64, labels map[string]string) error {
	if !c.limit.safeCheck(name, labels) {
		return fmt.Errorf("labels up to limit")
	}

	iface, ok := c.counters.Load(name)
	if !ok {
		return fmt.Errorf("metrics [%s] do not exists, create it first", name)
	}

	vec, ok := iface.(*prometheus.CounterVec)
	if !ok {
		c.counters.Delete(name)

		return fmt.Errorf("metrics [%s] is invalid, create it first", name)
	}

	vec.With(labels).Add(val)
	return nil
}

//CreateSummary create collector
func (c *PrometheusExporter) CreateSummary(opts *SummaryOpts) error {
	vec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: opts.Objectives,
	}, opts.Labels)

	if _, ok := c.summaries.Load(opts.Name); ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.summaries.Store(opts.Name, vec)

	prometheus.MustRegister(vec)
	return nil
}

//SummaryObserve set value
func (c *PrometheusExporter) SummaryObserve(name string, val float64, labels map[string]string) error {
	if !c.limit.safeCheck(name, labels) {
		return fmt.Errorf("labels up to limit")
	}

	iface, ok := c.summaries.Load(name)
	if !ok {
		return fmt.Errorf("metrics [%s] do not exists, create it first", name)
	}

	vec, ok := iface.(*prometheus.SummaryVec)
	if !ok {
		c.summaries.Delete(name)

		return fmt.Errorf("metrics [%s] is invalid, create it first", name)
	}

	vec.With(labels).Observe(val)
	return nil
}

//CreateHistogram create collector
func (c *PrometheusExporter) CreateHistogram(opts *HistogramOpts) error {
	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    opts.Name,
		Help:    opts.Help,
		Buckets: opts.Buckets,
	}, opts.Labels)

	if _, ok := c.histograms.Load(opts.Name); ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.histograms.Store(opts.Name, vec)

	prometheus.MustRegister(vec)
	return nil
}

//HistogramObserve set value
func (c *PrometheusExporter) HistogramObserve(name string, val float64, labels map[string]string) error {
	if !c.limit.safeCheck(name, labels) {
		return fmt.Errorf("labels up to limit")
	}

	iface, ok := c.histograms.Load(name)
	if !ok {
		return fmt.Errorf("metrics [%s] do not exists, create it first", name)
	}

	vec, ok := iface.(*prometheus.HistogramVec)
	if !ok {
		c.histograms.Delete(name)

		return fmt.Errorf("metrics [%s] is invalid, create it first", name)
	}

	vec.With(labels).Observe(val)
	return nil
}
