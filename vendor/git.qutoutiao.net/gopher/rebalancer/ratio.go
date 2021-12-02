package rebalancer

import "time"

type NewRatioFunc func() *ratio

// duration 10s
type ratio struct {
	duration time.Duration
	start    time.Time
}

func NewRatio(duration time.Duration) *ratio {
	return &ratio{
		duration: duration,
		start:    time.Now(),
	}
}

func (r *ratio) TargetRatio() float64 {
	// Here's why it's 0.5:
	// We are watching the following ratio
	// ratio = a / (a + d)
	// We can notice, that once we get to 0.5
	// 0.5 = a / (a + d)
	// we can evaluate that a = d
	// that means equilibrium, where we would allow all the requests
	// after this point to achieve ratio of 1 (that can never be reached unless d is 0)
	// so we stop from there
	multiplier := 0.5 / float64(r.duration)
	return multiplier * float64(time.Now().Sub(r.start))
}

// CalculateWeight 计算半开时的权重, 并判断边界条件
func (r *ratio) CalculateWeight(orgiWeight int) int {
	weight := int(r.TargetRatio() * float64(orgiWeight))

	if weight < 1 {
		weight = 1
	} else if weight > orgiWeight {
		weight = orgiWeight
	}

	return weight
}
