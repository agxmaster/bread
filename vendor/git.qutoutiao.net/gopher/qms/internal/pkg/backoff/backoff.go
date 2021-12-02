package backoff

import (
	"time"

	"github.com/cenkalti/backoff"
)

//constant for back off
const (
	BackoffJittered = "jittered"
	BackoffConstant = "constant"
	BackoffZero     = "zero"
	//DefaultBackOffKind is zero
	DefaultBackOffKind = BackoffZero
)

//GetBackOff return the the back off policy
//min and max unit is million second
func GetBackOff(kind string, min, max int) backoff.BackOff {
	switch kind {
	case BackoffJittered:
		return &backoff.ExponentialBackOff{
			InitialInterval:     time.Duration(min) * time.Millisecond,
			RandomizationFactor: backoff.DefaultRandomizationFactor,
			Multiplier:          backoff.DefaultMultiplier,
			MaxInterval:         time.Duration(max) * time.Millisecond,
			MaxElapsedTime:      0,
			Clock:               backoff.SystemClock,
		}
	case BackoffConstant:
		return backoff.NewConstantBackOff(time.Duration(min) * time.Millisecond)
	case BackoffZero:
		return &backoff.ZeroBackOff{}
	default:
		return &backoff.ZeroBackOff{}
	}

}
