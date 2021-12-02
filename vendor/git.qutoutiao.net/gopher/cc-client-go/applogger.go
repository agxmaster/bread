package cc

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/gopher/cc-client-go/proto-gen/admin_sdk"
)

type applogItem struct {
	variableTagId int64
	msg           string
}

type AppLogger struct {
	SDK *admin_sdk.SDKClient
	// key不存在的频率buffer
	freqBuffer map[int64]map[string]int64
	queue      chan applogItem
	closed     chan struct{}
	options    *ApploggerOptions
	center     *ConfigCenter
}

func (l *AppLogger) SetCenter(center *ConfigCenter) {
	l.center = center
}

func (l AppLogger) process() {
	timer := time.NewTicker(l.options.bufferFlushInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			l.Flush()
		case item := <-l.queue:
			if l.freqBuffer[item.variableTagId] == nil {
				l.freqBuffer[item.variableTagId] = make(map[string]int64)
			}
			l.freqBuffer[item.variableTagId][item.msg]++
		case <-l.closed:
			return
		}
	}
}

func NewAppLogger(center *ConfigCenter, SDK *admin_sdk.SDKClient, opts ...AppLoggerOption) *AppLogger {
	options := &ApploggerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.queueSize == 0 {
		options.queueSize = DefaultQueueSize
	}
	if options.bufferFlushInterval == 0 {
		options.bufferFlushInterval = DefaultFlushInterval
	}
	l := &AppLogger{center: center, SDK: SDK, options: options, queue: make(chan applogItem, options.queueSize), freqBuffer: make(map[int64]map[string]int64)}
	go l.process()
	return l
}

func (l AppLogger) Flush() error {
	var in = &admin_sdk.GetKeyNotFoundReq{}
	in.Client = l.SDK
	for variableTagId, freq := range l.freqBuffer {
		for key, count := range freq {
			in.KeyErrors = append(in.KeyErrors, &admin_sdk.KeyNotFound{Key: key, Count: count, ConfigVariableTagId: variableTagId})
		}
	}
	if len(in.KeyErrors) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := l.center.getClient().GetKeyNotFound(ctx, in)
	l.resetBuffer()
	return err
}

func (l AppLogger) resetBuffer() {
	for k := range l.freqBuffer {
		delete(l.freqBuffer, k)
	}
}

func (l AppLogger) Error(msg string) {
	select {
	case l.queue <- applogItem{
		variableTagId: atomic.LoadInt64(&l.center.latestConfigVariableTagId),
		msg:           msg,
	}:
	default:
	}
}

func (l AppLogger) Errorf(msg string, args ...interface{}) {
	l.Error(fmt.Sprintf(msg, args...))
}

func (l AppLogger) Info(msg string) {
	return
}

func (l AppLogger) Infof(msg string, args ...interface{}) {
	l.Info(fmt.Sprintf(msg, args...))
}

func (l AppLogger) Close() error {
	select {
	case l.closed <- struct{}{}:
	default:
	}
	return l.Flush()
}
