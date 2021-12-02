package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func IncOp(labels prometheus.Labels) {
	gormCounter.With(labels).Inc()
}

func ObserveOp(labels prometheus.Labels, startedAt time.Time) {
	gormLatency.With(labels).Observe(time.Since(startedAt).Seconds())
}
