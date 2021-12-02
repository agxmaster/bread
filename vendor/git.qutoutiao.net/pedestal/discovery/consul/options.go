package consul

import (
	"time"

	"git.qutoutiao.net/pedestal/discovery/cache"
)

type option struct {
	stale      bool
	agentCache bool

	watchTimeout time.Duration
	watchLatency time.Duration

	// degrade with passingOnly=false settings
	passingOnly bool
	threshold   float32
	emergency   float32

	// local file cache interface
	cache             cache.Interface
	cacheSyncInterval time.Duration

	calmInterval time.Duration
	useCatalog   bool
	debug        bool
}

func (o *option) enableDegrade() bool {
	return o.threshold > 0
}

type ConsulOption func(*option)

// 强制使用 catalog 接口
func WithUseCatalog(isCatalog bool) ConsulOption {
	return func(o *option) {
		o.useCatalog = isCatalog
	}
}

// enable consul api stale
func WithStale(stale bool) ConsulOption {
	return func(o *option) {
		o.stale = stale
	}
}

// enable consul agent cache
func WithAgentCache(cache bool) ConsulOption {
	return func(o *option) {
		o.agentCache = cache
	}
}

// filter with passingOnly=true
func WithPassingOnly(passingOnly bool) ConsulOption {
	return func(o *option) {
		o.passingOnly = passingOnly
	}
}

// set threshold of degrade calc with (100 * passing / total), valid values are [0.0-1.0], default to 0, mains no degrade.
//
// NOTE: only for passingOnly=true!
func WithDegrade(threshold float32) ConsulOption {
	return func(o *option) {
		o.threshold = threshold
	}
}

// set threshold of emergency calc with (100 * critical / total), valid values are [0.0-100.0], default to 80.
//
// NOTE: only for passingOnly=true!
func WithEmergency(threshold float32) ConsulOption {
	return func(o *option) {
		o.emergency = threshold
	}
}

// set custom cache of service resolver
func WithCache(cacher cache.Interface) ConsulOption {
	return func(o *option) {
		o.cache = cacher
	}
}

// set interval of dump to local file for emergency.
func WithDumpInterval(interval time.Duration) ConsulOption {
	return func(o *option) {
		o.cacheSyncInterval = interval
	}
}

// WithCalmInterval 设置阀值监控时间
func WithCalmInterval(interval time.Duration) ConsulOption {
	return func(o *option) {
		o.calmInterval = interval
	}
}

// WithWatchTimeout wait >= 30s and wait <= 10m
func WithWatchTimeout(timeout time.Duration) ConsulOption {
	return func(o *option) {
		if timeout.Seconds() <= 0 {
			timeout = DefaultServiceWatchTimeout
		}

		if timeout.Minutes() > 10 {
			timeout = 10 * time.Minute
		}

		o.watchTimeout = timeout
	}
}

// WithWatchMaxLatency latency >= 3s and latency <= 30s
func WithWatchMaxLatency(latency time.Duration) ConsulOption {
	return func(o *option) {
		if latency.Seconds() <= 0 {
			latency = DefaultServiceWatchLatency
		}

		if latency.Minutes() > 1 {
			latency = time.Minute
		}

		o.watchLatency = latency
	}
}

func WithDebug(debug bool) ConsulOption {
	return func(o *option) {
		o.debug = debug
	}
}
