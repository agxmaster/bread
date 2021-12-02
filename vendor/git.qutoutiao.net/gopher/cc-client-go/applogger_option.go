package cc

import (
	"time"
)

type AppLoggerOption func(options *ApploggerOptions)

type ApploggerOptions struct {
	// flush间隔
	bufferFlushInterval time.Duration
	// channel buffer 大小
	queueSize int
}

func AppBufferFlushInterval(bufferFlushInterval time.Duration) AppLoggerOption {
	return func(options *ApploggerOptions) {
		options.bufferFlushInterval = bufferFlushInterval
	}
}

func AppQueueSize(queueSize int) AppLoggerOption {
	return func(options *ApploggerOptions) {
		options.queueSize = queueSize
	}
}
