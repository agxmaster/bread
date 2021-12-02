package util

import (
	"errors"
	"math/rand"
	"time"
)

// Jitter returns a time.Duration between duration and duration + maxFactor *
// duration.
//
// This allows clients to avoid converging on periodic behavior. If maxFactor
// is 0.0, a suggested default value will be chosen.
func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

// JitterUntil loops until stop channel is closed, running f every period.
//
// If jitterFactor is positive, the period is jittered before every run of f.
// If jitterFactor is not positive, the period is unchanged and not jittered.
//
// If sliding is true, the period is computed after f runs. If it is false then
// period includes the runtime for f.
//
// Close stopCh to stop. f may not be invoked if stop channel is already
// closed. Pass NeverStop to if you don't want it stop.
func JitterUntil(f func(), period time.Duration, jitterFactor float64, sliding bool, stopCh <-chan struct{}) {
	var t *time.Timer
	var sawTimeout bool

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		jitteredPeriod := period
		if jitterFactor > 0.0 {
			jitteredPeriod = Jitter(period, jitterFactor)
		}

		if !sliding {
			t = ResetOrReuseTimer(t, jitteredPeriod, sawTimeout)
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					return
				}
			}()

			f()
		}()

		if sliding {
			t = ResetOrReuseTimer(t, jitteredPeriod, sawTimeout)
		}

		// NOTE: b/c there is no priority selection in golang
		// it is possible for this to race, meaning we could
		// trigger t.C and stopCh, and t.C select falls through.
		// In order to mitigate we re-check stopCh at the beginning
		// of every loop to prevent extra executions of f().
		select {
		case <-stopCh:
			return
		case <-t.C:
			sawTimeout = true
		}
	}
}

// JitterTimeout loops until timeout, running f every period.
//
// See JitterUntil for more details.
func JitterTimeout(f func() error, period, timeout time.Duration, jitterFactor float64, sliding bool) (err error) {
	stopCh := make(chan struct{})

	timer := time.AfterFunc(timeout, func() {
		close(stopCh)
	})
	defer timer.Stop()

	JitterUntil(func() {
		select {
		case <-timer.C:
			err = errors.New("timeout")
		default:
			err = f()
			if err == nil {
				close(stopCh)
			}
		}
	}, period, jitterFactor, sliding, stopCh)

	return
}

// ResetOrReuseTimer avoids allocating a new timer if one is already in use.
// Not safe for multiple threads.
func ResetOrReuseTimer(t *time.Timer, d time.Duration, sawTimeout bool) *time.Timer {
	if t == nil {
		return time.NewTimer(d)
	}

	if !t.Stop() && !sawTimeout {
		<-t.C
	}

	t.Reset(d)
	return t
}
