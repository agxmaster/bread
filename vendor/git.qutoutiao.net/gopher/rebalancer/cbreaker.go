// Package gobreaker implements the Circuit Breaker pattern.
// See https://msdn.microsoft.com/en-us/library/dn589784.aspx.
package rebalancer

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/rebalancer/degrade"
	"github.com/vulcand/oxy/memmetrics"
	"golang.org/x/time/rate"
)

const (
	counterBuckets           = 10
	counterResolution        = time.Second
	defaultCheckPeriod       = 100 * time.Millisecond
	defaultFallbackDuration  = 10 * time.Second
	defaultRecoveryDuration  = 10 * time.Second
	defaultCBreakerThreshold = 80

	defaultRate                 = 100
	defaultBreakerOpenValue     = 2 // 初始metric值
	defaultBreakerHalfOpenValue = 3 // 初始metric值
)

type NewCBreakerFunc func(...CBreakerOption) (*cbreaker, error)

type CBreakerOption func(*Settings)

func CBreakerWithAddress(address string) CBreakerOption {
	return func(settings *Settings) {
		settings.Name = address
	}
}

func CBreakerDisabled() CBreakerOption {
	return func(settings *Settings) {
		settings.Disabled = true
	}
}

type Settings struct {
	Name              string
	CheckPeriod       time.Duration
	FallbackDuration  time.Duration
	RecoveryDuration  time.Duration
	ReadyToTrip       func(counts *Counts) bool
	IsInterceptChange func(name string, from, to State) bool
	OnStateChange     func(name string, from, to State)
	Disabled          bool
	Counts            *Counts
	LimiterRate       float64
}

// cbreaker is a state machine to prevent sending requests that are likely to fail.
type cbreaker struct {
	name              string
	solutionDuration  time.Duration // 计算超时时间，默认10s
	checkPeriod       time.Duration // 统计周期，默认100ms，到期counts清0
	fallbackDuration  time.Duration // open超时时间，默认10s
	recoveryDuration  time.Duration // recovery时间，默认10s
	readyToTrip       func(counts *Counts) bool
	isInterceptChange func(name string, from, to State) bool
	onStateChange     func(name string, from, to State)

	mutex     sync.Mutex
	state     State
	counts    *Counts
	expiry    time.Time
	recovery  time.Time
	limiter   *rate.Limiter
	timestamp time.Time
}

func NewCBreaker(st *Settings) (cb *cbreaker, err error) {
	if st.Disabled {
		return nil, nil
	}

	cb = new(cbreaker)

	cb.name = st.Name
	cb.onStateChange = st.OnStateChange

	if st.IsInterceptChange == nil {
		cb.isInterceptChange = func(name string, from, to State) bool {
			return false
		}
	} else {
		cb.isInterceptChange = st.IsInterceptChange
	}

	if st.CheckPeriod == 0 {
		cb.checkPeriod = defaultCheckPeriod
	} else {
		cb.checkPeriod = st.CheckPeriod
	}

	if st.FallbackDuration == 0 {
		cb.fallbackDuration = defaultFallbackDuration
	} else {
		cb.fallbackDuration = st.FallbackDuration
	}

	if st.RecoveryDuration == 0 {
		cb.recoveryDuration = defaultRecoveryDuration
	} else {
		cb.recoveryDuration = st.RecoveryDuration
	}

	if st.ReadyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	} else {
		cb.readyToTrip = st.ReadyToTrip
	}

	if st.Counts == nil {
		cb.counts, err = NewCounts()
		if err != nil {
			return nil, err
		}
	} else {
		cb.counts = st.Counts
	}

	if st.LimiterRate > 0 {
		cb.limiter = rate.NewLimiter(rate.Limit(st.LimiterRate), int(st.LimiterRate))
	} else {
		cb.limiter = rate.NewLimiter(rate.Limit(defaultRate), defaultRate)
	}

	cb.toNewGeneration(time.Now())

	return
}

// State returns the current state of the CircuitBreaker.
func (cb *cbreaker) State() State {
	if cb != nil {
		cb.mutex.Lock()
		defer cb.mutex.Unlock()

		return cb.state
	}
	return StateClosed
}

// Record 记录统计信息。
// 状态变更：
// closed: 如果满足ReadyToTrip条件，状态变更为open；
// half_open: 如果Recovery时间内没有错误，状态变更为closed；有任意错误则状态变更为open。
func (cb *cbreaker) Record(success bool) {
	if cb == nil {
		return
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// 1. 获取当前状态
	now := time.Now()
	state := cb.currentState(now)

	switch state {
	case StateOpen:
		if !success {
			// metrics
			degrade.EventSet(degrade.EventBreakerOpenStatus, cb.name, now.Sub(cb.timestamp).Seconds()+defaultBreakerOpenValue)
		}
		return
	case StateHalfOpen:
		if !success {
			// metrics
			degrade.EventSet(degrade.EventBreakerHalfOpenStatus, cb.name, now.Sub(cb.timestamp).Seconds()+defaultBreakerHalfOpenValue)
		}
		fallthrough
	default: // half_open、closed
		if cb.limiter.Allow() {
			cb.counts.onRequest()
			if success {
				cb.counts.onSuccess()
			} else {
				cb.counts.onFailure()
			}
		}
		cb.checkAndSet(now)
	}
}

func (cb *cbreaker) checkAndSet(now time.Time) {
	if !cb.expiry.Before(now) {
		return
	}

	cb.expiry = now.Add(cb.checkPeriod) // 更新check 超时时间
	if cb.readyToTrip(cb.counts) {
		cb.setState(StateOpen, now)
	}
}

func (cb *cbreaker) currentState(now time.Time) State {
	switch cb.state {
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	case StateHalfOpen:
		if cb.recovery.Before(now) {
			cb.setState(StateClosed, now)
		}
	}
	return cb.state
}

// setState 状态变更
func (cb *cbreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	// before
	prev := cb.state
	if cb.isInterceptChange(cb.name, prev, state) {
		return
	}

	// process
	cb.state = state
	if prev == StateClosed && state == StateOpen {
		cb.timestamp = now
	}
	cb.toNewGeneration(now)

	// after
	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

func (cb *cbreaker) toNewGeneration(now time.Time) {
	// clear
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case StateOpen:
		if cb.fallbackDuration == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.fallbackDuration)
		}
		degrade.EventDeleteLabelValues(degrade.EventBreakerHalfOpenStatus, cb.name)
		degrade.EventSet(degrade.EventBreakerOpenStatus, cb.name, now.Sub(cb.timestamp).Seconds()+defaultBreakerOpenValue)
	case StateHalfOpen:
		if cb.recoveryDuration == 0 {
			cb.recovery = zero
		} else {
			cb.recovery = now.Add(cb.recoveryDuration)
		}
		degrade.EventDeleteLabelValues(degrade.EventBreakerOpenStatus, cb.name)
		degrade.EventSet(degrade.EventBreakerHalfOpenStatus, cb.name, now.Sub(cb.timestamp).Seconds()+defaultBreakerHalfOpenValue)
		//fallthrough
		if cb.checkPeriod == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.checkPeriod)
		}
	case StateClosed:
		if cb.checkPeriod == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.checkPeriod)
		}
		degrade.EventDeleteLabelValues(degrade.EventBreakerOpenStatus, cb.name)
		degrade.EventDeleteLabelValues(degrade.EventBreakerHalfOpenStatus, cb.name)
	}
}

// State is a type that represents a state of cbreaker.
type State int

// These constants are states of cbreaker.
const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var (
	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests
	ErrTooManyRequests = errors.New("too many requests")
	// ErrOpenState is returned when the CB state is open
	ErrOpenState = errors.New("circuit breaker is open")
)

// String implements stringer interface.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}

// Counts holds the numbers of requests and their successes/failures.
// cbreaker clears the internal Counts either
// on the change of the state or at the closed-state intervals.
// Counts ignores the results of the requests sent before clearing.
type Counts struct {
	Requests       *memmetrics.RollingCounter
	TotalSuccesses *memmetrics.RollingCounter
	TotalFailures  *memmetrics.RollingCounter
}

func NewCounts() (*Counts, error) {
	newCounter := func() (*memmetrics.RollingCounter, error) {
		return memmetrics.NewCounter(counterBuckets, counterResolution)
	}
	requests, err := newCounter()
	if err != nil {
		return nil, err
	}
	totalSuccess, err := newCounter()
	if err != nil {
		return nil, err
	}
	totalFailures, err := newCounter()
	if err != nil {
		return nil, err
	}
	return &Counts{
		Requests:       requests,
		TotalSuccesses: totalSuccess,
		TotalFailures:  totalFailures,
	}, nil
}

func (c *Counts) onRequest() {
	c.Requests.Inc(1)
}

func (c *Counts) onSuccess() {
	c.TotalSuccesses.Inc(1)
}

func (c *Counts) onFailure() {
	c.TotalFailures.Inc(1)
}

func (c *Counts) clear() {
	c.Requests.Reset()
	c.TotalSuccesses.Reset()
	c.TotalFailures.Reset()
}

// defaultReadyToTrip 错误率大于90%，则为true
func defaultReadyToTrip(counts *Counts) bool {
	if counts.Requests.Count() == 0 { // 没有请求
		return false
	}
	return float64(counts.TotalFailures.Count())/float64(counts.Requests.Count())*100 > 90
}
