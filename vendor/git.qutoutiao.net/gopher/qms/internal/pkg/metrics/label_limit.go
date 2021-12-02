package metrics

import (
	"strings"
	"sync"
	"sync/atomic"
)

// MaxLabelLimit 表示metrics输出的最大行数
const MaxLabelLimit = 50000

type labelLimit struct {
	numLabels uint32
	allLabels sync.Map
}

func newLabelLimit() *labelLimit {
	return &labelLimit{
		allLabels: sync.Map{},
	}
}
func (l *labelLimit) safeCheck(name string, labels map[string]string) bool {
	sb := strings.Builder{}
	sb.WriteString(name)
	for k, v := range labels {
		sb.WriteString(k)
		sb.WriteString(v)
	}
	key := sb.String()

	if _, ok := l.allLabels.Load(key); ok {
		return true
	}

	if atomic.LoadUint32(&l.numLabels) >= MaxLabelLimit {
		return false
	}

	atomic.AddUint32(&l.numLabels, 1)
	l.allLabels.Store(key, struct{}{})

	return true
}
