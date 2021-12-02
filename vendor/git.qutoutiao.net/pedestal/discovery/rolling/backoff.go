package rolling

import (
	"os"
	"sync"
	"time"

	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/util"
)

type Backoff struct {
	mux      sync.Mutex
	window   *Window
	delay    time.Duration
	timeout  time.Duration
	latency  time.Duration
	issuedAt time.Time
	logger   logger.Interface

	// for shift
	shifti   int
	shiftn   int
	shiftd   time.Duration
	shifting bool
}

func NewBackoff(timeout time.Duration) *Backoff {
	return NewBackoffWithLogger(timeout, DefaultLatency, logger.New(os.Stderr))
}

func NewBackoffWithLogger(timeout, latency time.Duration, logger logger.Interface) *Backoff {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if latency <= 0 {
		latency = DefaultLatency
	}

	return &Backoff{
		window:  NewWindow(DefaultSize),
		delay:   0,
		shifti:  3,
		shiftd:  3 * time.Second,
		timeout: timeout,
		latency: latency,
		logger:  logger,
	}
}

func (b *Backoff) Delay(t, key string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.window.Append(time.Now())

	delay := b.calcDelay()
	if delay > 0 {
		if delay > b.latency {
			delay = b.latency
		}

		if b.logger != nil {
			b.logger.Warnf("%s backoff of %s with %v", t, key, delay)
		}

		time.Sleep(delay)
	}
}

func (b *Backoff) DelayWithShift(t, key string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.window.Append(time.Now())

	delay := b.shitDelay()
	if delay > 0 {
		if b.logger != nil {
			b.logger.Warnf("%s backoff of %s with %v (shift{i=%d,n=%d)\n", t, key, delay, b.shifti, b.shiftn)
		}

		time.Sleep(delay)
	}
}

func (b *Backoff) calcDelay() time.Duration {
	// no delay required
	if b.issuedAt.IsZero() || time.Since(b.issuedAt) > b.timeout {
		b.issuedAt = time.Time{}
		b.delay = 0
	}

	seconds := 1
	max := 2

	// 2 changes within 1s = 2qps
	if b.window.Match(seconds, max) {
		b.delay = 1 * time.Second
	}

	seconds = 3
	max = 3

	// 3 changes within 3s = 1qps
	if b.window.Match(seconds, max) {
		b.delay = 2 * time.Second
	}

	seconds = 6
	max = 4

	// 4 changes within 6s = 0.75qps
	if b.window.Match(seconds, max) {
		b.delay = 3 * time.Second
	}

	seconds = 10
	max = 6

	// 6 changes within 10s = 0.8qps
	if b.window.Match(seconds, max) {
		b.delay = 5 * time.Second

	}

	if b.delay > 0 {
		defer func() {
			b.delay = b.delay * time.Duration(max) / time.Duration(seconds)
		}()

		b.issuedAt = time.Now()
	}

	return b.delay
}

func (b *Backoff) shitDelay() time.Duration {
	delay := b.calcDelay()

	// only shift for 5s delay with more than 3 times
	if !b.shifting && delay >= b.shiftd {
		b.shiftn++
		b.shifting = b.shiftn > b.shifti
	}

	if b.shifting && b.shiftn > 0 {
		delay += util.Jitter(delay, float64(b.shiftn))

		if delay < b.shiftd {
			delay = b.shiftd
		}

		b.shiftn--
		b.shifting = b.shiftn > 0
	}

	if delay > b.latency {
		delay = b.latency
	}

	return delay
}
