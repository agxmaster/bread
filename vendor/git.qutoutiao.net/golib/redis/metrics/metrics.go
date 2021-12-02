package metrics

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

func IncCmd(labels prometheus.Labels) {
	redisCounter.With(labels).Inc()
}

func SetConnPool(addr string, stats *redis.PoolStats) {
	labels := prometheus.Labels{
		"to": addr,
	}

	labels["stats"] = "total"
	redisConnPoolCounter.With(labels).Set(float64(stats.TotalConns))

	labels["stats"] = "idle"
	redisConnPoolCounter.With(labels).Set(float64(stats.IdleConns))

	labels["stats"] = "stale"
	redisConnPoolCounter.With(labels).Set(float64(stats.StaleConns))

	labels["stats"] = "hits"
	redisConnPoolCounter.With(labels).Set(float64(stats.Hits))

	labels["stats"] = "misses"
	redisConnPoolCounter.With(labels).Set(float64(stats.Misses))

	labels["stats"] = "timeouts"
	redisConnPoolCounter.With(labels).Set(float64(stats.Timeouts))
}

func ObserveCmd(labels prometheus.Labels, startedAt time.Time) {
	redisLatency.With(labels).Observe(time.Since(startedAt).Seconds())
}
