package sentinel

import (
	"time"
)

type option struct {
	// 多久未请求则删除 watchKey
	cleanInterval time.Duration

	// 多久批量获取一次 watchKeys
	fetchInterval time.Duration
}

type Option func(o *option)

// 多久清空一次 watch
func WithCleanInterval(dur time.Duration) Option {
	return func(o *option) {
		o.cleanInterval = dur
	}
}

// 多久获取一次数据
func WithFetchInterval(dur time.Duration) Option {
	return func(o *option) {
		o.fetchInterval = dur
	}
}
