package metrics

import (
	"time"

	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var (
	registries      = make(map[string]NewRegistry)
	defaultRegistry Registry
)

//CreateGauge init a new gauge type
func CreateGauge(opts *GaugeOpts) error {
	return defaultRegistry.CreateGauge(opts)
}

//CreateCounter init a new counter type
func CreateCounter(opts *CounterOpts) error {
	return defaultRegistry.CreateCounter(opts)
}

//CreateSummary init a new summary type
func CreateSummary(opts *SummaryOpts) error {
	return defaultRegistry.CreateSummary(opts)
}

//CreateHistogram init a new summary type
func CreateHistogram(opts *HistogramOpts) error {
	return defaultRegistry.CreateHistogram(opts)
}

//GaugeSet set a new value to a collector
func GaugeSet(name string, val float64, labels map[string]string) error {
	return defaultRegistry.GaugeSet(name, val, labels)
}

//GaugeReset reset value
func GaugeReset(name string) error {
	return defaultRegistry.GaugeReset(name)
}

//CounterAdd increase value of a collector
func CounterAdd(name string, val float64, labels map[string]string) error {
	return defaultRegistry.CounterAdd(name, val, labels)
}

//SummaryObserve gives a value to summary collector
func SummaryObserve(name string, val float64, labels map[string]string) error {
	return defaultRegistry.SummaryObserve(name, val, labels)
}

//HistogramObserve gives a value to histogram collector
func HistogramObserve(name string, val float64, labels map[string]string) error {
	return defaultRegistry.HistogramObserve(name, val, labels)
}

//CounterOpts is options to create a counter options
type CounterOpts struct {
	Name   string
	Help   string
	Labels []string
}

//GaugeOpts is options to create a gauge collector
type GaugeOpts struct {
	Name   string
	Help   string
	Labels []string
}

//SummaryOpts is options to create summary collector
type SummaryOpts struct {
	Name       string
	Help       string
	Labels     []string
	Objectives map[float64]float64
}

//HistogramOpts is options to create histogram collector
type HistogramOpts struct {
	Name    string
	Help    string
	Labels  []string
	Buckets []float64
}

// 自定义metric标签
type CustomLabel struct {
	LabelName     string //标签key
	LabelValueKey string //label取值对应的key (通过middleware添加到ginContext.Keys中)
}

//Options control config
type Options struct {
	FlushInterval time.Duration
}

//InstallPlugin install metrics registry
func InstallPlugin(name string, f NewRegistry) {
	if name != "prometheus" {
		qlog.Warnf("only prometheus plugin is supported!")
	}

	registries[name] = f
}
