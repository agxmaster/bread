package rolling

import (
	"sync"
	"time"
)

// rolling window
type Window struct {
	mux  sync.Mutex
	size int
	data []int
	now  time.Time
	cur  int
}

func NewWindow(size int) *Window {
	return &Window{
		size: size,
		data: make([]int, size),
		cur:  -1,
		now:  time.Now(),
	}
}

func (win *Window) Append(now time.Time) {
	win.mux.Lock()
	defer win.mux.Unlock()

	since := int(now.Unix() - win.now.Unix())

	// 当前时间窗口
	if since == 0 {
		if win.cur < 0 {
			win.cur = (win.cur + win.size) % win.size
		}

		win.data[win.cur]++
		return
	}

	// 时间已经滚动了很多秒
	if since > win.size {
		since = win.size
	}

	// 窗口滚动
	for i := 0; i < since; i++ {
		win.cur = (win.cur + 1) % win.size
		win.data[win.cur] = 0
	}

	win.data[win.cur]++
	win.now = now
}

// 连续 seconds 秒的请求数达到 max 次
func (win *Window) Match(seconds int, max int) bool {
	win.mux.Lock()
	defer win.mux.Unlock()

	if seconds > win.size {
		seconds = win.size
	}

	last := (win.cur - seconds + win.size + 1) % win.size

	count := 0
	for {
		count += win.data[last]
		if count >= max {
			return true
		}

		if last == win.cur {
			break
		}

		last = (last + 1) % win.size
	}

	return false
}

func (win *Window) Total() int {
	win.mux.Lock()
	defer win.mux.Unlock()

	total := 0
	for i := range win.data {
		total += win.data[i]
	}

	return total
}
