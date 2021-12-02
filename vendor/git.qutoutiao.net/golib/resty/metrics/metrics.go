package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func IncResolverTotals(labels prometheus.Labels) {
	resolverTotalsCounter.With(labels).Inc()
}

func IncResolverFailures(labels prometheus.Labels) {
	resolverFailuresCounter.With(labels).Inc()
}

func IncNetFailures(labels prometheus.Labels) {
	netFailuresCounter.With(labels).Inc()
}

func IncRequestTotals(labels prometheus.Labels) {
	requestTotalsCounter.With(labels).Inc()
}

func IncRequestFailures(labels prometheus.Labels) {
	requestFailuresCounter.With(labels).Inc()
}

func ObserveRequest(labels prometheus.Labels, startedAt time.Time) {
	requestLatency.With(labels).Observe(time.Since(startedAt).Seconds())
}
