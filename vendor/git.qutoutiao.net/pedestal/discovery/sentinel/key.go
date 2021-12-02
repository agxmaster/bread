package sentinel

import (
	"sync"
	"time"
)

type watchKey struct {
	lock    sync.RWMutex
	dc      string
	name    string
	tags    []string
	last    time.Time
	enabled bool
}

func (w *watchKey) setSentinel(b bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.enabled = b
}

func (w *watchKey) useSentinel() bool {
	if w == nil {
		return false
	}

	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.enabled
}

func (w *watchKey) setLastTime(last time.Time) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.last = last
}

func (w *watchKey) isExpired(duration time.Duration) bool {
	if w == nil {
		return false
	}

	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.last.Add(duration).After(time.Now())
}
