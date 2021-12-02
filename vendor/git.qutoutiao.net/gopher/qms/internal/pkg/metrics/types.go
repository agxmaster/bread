package metrics

//Registry holds all of metrics collectors
//name is a unique ID for different type of metrics
type Registry interface {
	CreateGauge(opts *GaugeOpts) error
	CreateCounter(opts *CounterOpts) error
	CreateSummary(opts *SummaryOpts) error
	CreateHistogram(opts *HistogramOpts) error

	GaugeSet(name string, val float64, labels map[string]string) error
	GaugeReset(name string) error
	CounterAdd(name string, val float64, labels map[string]string) error
	SummaryObserve(name string, val float64, Labels map[string]string) error
	HistogramObserve(name string, val float64, labels map[string]string) error
}

//NewRegistry create a registry
type NewRegistry func(opts Options) Registry
