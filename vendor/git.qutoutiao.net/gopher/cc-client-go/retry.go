package cc

import (
	"math/rand"
	"time"
)

// jitterUp adds random jitter to the duration.
//
// This adds or subtracts time from the duration within a given jitter fraction.
// For example for 10s and jitter 0.1, it will return a time within [9s, 11s])
//
func jitterUp(duration time.Duration, jitter float64) time.Duration {
	multiplier := jitter * (rand.Float64()*2 - 1)
	return time.Duration(float64(duration) * (1 + multiplier))
}

// For example waitBetween=1s and jitter=0.10 can generate waits between 900ms and 1100ms.
func backoffExponentialWithJitter(jitterFraction float64, initInterval, maxInterval time.Duration) backoffFunc {
	return func(attempt uint64) time.Duration {
		if attempt == 0 {
			return 0
		}
		interval := (1 << (attempt - 1)) * initInterval
		if interval > maxInterval {
			interval = maxInterval
		}
		return jitterUp(interval, jitterFraction)
	}
}

// backoffFunc denotes a family of functions that control the backoff duration between call retries.
//
// They are called with an identifier of the attempt, and should return a time the system client should
// hold off for. If the time returned is longer than the `context.Context.Deadline` of the request
// the deadline of the request takes precedence and the wait will be interrupted before proceeding
// with the next iteration.
type backoffFunc func(attempt uint64) time.Duration
