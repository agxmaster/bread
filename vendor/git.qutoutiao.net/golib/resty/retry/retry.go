package retry

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/pedestal/discovery/util"
)

const (
	DefaultTimeout = 5 * time.Second
	DefaultPeriod  = 300 * time.Millisecond
	DefaultFactor  = 1.0
)

var (
	ErrRetrying = errors.New("retrying")
)

func JitterRetry(fn func(retries uint32) error, maxRetries int, period time.Duration, factor float64) {
	if period <= 0 {
		period = DefaultPeriod
	}
	if factor <= 0 {
		factor = DefaultFactor
	}

	var counter uint32

	stopCh := make(chan struct{})
	util.JitterUntil(func() {
		retries := atomic.AddUint32(&counter, 1)
		if retries > uint32(maxRetries) {
			close(stopCh)
			return
		}

		err := fn(retries)
		if err == nil {
			close(stopCh)
		}
	}, period, factor, true, stopCh)
}

func TimeoutRetry(fn func(retries uint32) error, maxRetries int, period, timeout time.Duration) (err error) {
	if period <= 0 {
		period = DefaultPeriod
	}
	factor := DefaultFactor

	var counter uint32

	stopCh := make(chan struct{})
	timer := time.AfterFunc(timeout, func() {
		err = fmt.Errorf("retry with timeout after %v, %d retries", timeout, atomic.LoadUint32(&counter))

		close(stopCh)
	})
	defer timer.Stop()

	util.JitterUntil(func() {
		retries := atomic.AddUint32(&counter, 1)
		if retries > uint32(maxRetries) {
			close(stopCh)
			return
		}

		err := fn(retries)
		if err == nil {
			close(stopCh)
		}
	}, period, factor, true, stopCh)

	return
}
