package cc

import (
	"context"
	"fmt"
	"time"

	"git.qutoutiao.net/gopher/cc-client-go/proto-gen/admin_sdk"
)

type item struct {
	level admin_sdk.LogLevel
	msg   string
}

type Logger struct {
	SDK    *admin_sdk.SDKClient
	center *ConfigCenter
	// 日志buffer
	logBuffer []item
	queue     chan item
	closed    chan struct{}
	options   *diagloggerOptions
}

func (l Logger) process() {
	timer := time.NewTicker(l.options.bufferFlushInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			l.Flush()
		case it := <-l.queue:
			l.logBuffer = append(l.logBuffer, it)
			if len(l.logBuffer) >= l.options.maxLogSize {
				l.Flush()
			}
		case <-l.closed:
			return
		}
	}
}

func NewLogger(center *ConfigCenter, SDK *admin_sdk.SDKClient, opts ...LoggerOption) *Logger {
	options := &diagloggerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.queueSize == 0 {
		options.queueSize = DefaultQueueSize
	}
	if options.maxLogSize == 0 {
		options.maxLogSize = DefaultMaxLogSize
	}
	if options.bufferFlushInterval == 0 {
		options.bufferFlushInterval = DefaultFlushInterval
	}
	l := &Logger{center: center, SDK: SDK, options: options, queue: make(chan item, options.queueSize)}
	go l.process()
	return l
}

func (l *Logger) Flush() error {
	if len(l.logBuffer) == 0 {
		return nil
	}
	var logs []*admin_sdk.Log
	timestamp := time.Now().Unix()
	for _, item := range l.logBuffer {
		logs = append(logs, &admin_sdk.Log{
			Level:     item.level,
			Msg:       item.msg,
			Timestamp: timestamp})
	}
	in := &admin_sdk.LogReq{
		Client: l.SDK,
		Logs:   logs,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := l.center.getClient().Log(ctx, in)
	l.resetBuffer()
	return err
}

func (l *Logger) resetBuffer() {
	l.logBuffer = l.logBuffer[:0]
}

func (l Logger) Error(msg string) {
	select {
	case l.queue <- item{level: admin_sdk.LogLevel_ERROR, msg: msg}:
	default:
	}
}

func (l Logger) Errorf(msg string, args ...interface{}) {
	l.Error(fmt.Sprintf(msg, args...))
}

func (l Logger) Info(msg string) {
	select {
	case l.queue <- item{level: admin_sdk.LogLevel_INFO, msg: msg}:
	default:
	}
	return
}

func (l Logger) Infof(msg string, args ...interface{}) {
	l.Info(fmt.Sprintf(msg, args...))
}

func (l Logger) Close() error {
	select {
	case l.closed <- struct{}{}:
	default:
	}
	return l.Flush()
}
