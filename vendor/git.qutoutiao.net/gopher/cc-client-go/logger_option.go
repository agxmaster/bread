package cc

import (
	"time"
)

type diagloggerOptions struct {
	// flush间隔
	bufferFlushInterval time.Duration
	// 最大日志大小
	maxLogSize int
	// channel buffer 大小
	queueSize int
}

const (
	DefaultQueueSize     = 1000
	DefaultMaxLogSize    = 1000
	DefaultFlushInterval = 10 * time.Second
)

type LoggerOption func(options *diagloggerOptions)

func BufferFlushInterval(bufferFlushInterval time.Duration) LoggerOption {
	return func(options *diagloggerOptions) {
		options.bufferFlushInterval = bufferFlushInterval
	}
}

func MaxLogSize(maxLogSize int) LoggerOption {
	return func(options *diagloggerOptions) {
		options.maxLogSize = maxLogSize
	}
}

func QueueSize(queueSize int) LoggerOption {
	return func(options *diagloggerOptions) {
		options.queueSize = queueSize
	}
}
