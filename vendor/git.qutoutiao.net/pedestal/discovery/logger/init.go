package logger

import (
	"log"
	"os"
	"sync"

	"git.qutoutiao.net/pedestal/discovery/logger/hclog"
)

var (
	globalMux    sync.Mutex
	globalLogger Interface = NewWithHclog(hclog.NewWithSkipFrameCount(6, os.Stderr))
)

func SetLogger(nlog Interface) {
	globalMux.Lock()
	if nlog != nil {
		globalLogger = nlog
	}
	globalMux.Unlock()
}

func Errorf(format string, v ...interface{}) {
	if globalLogger == nil {
		log.Printf(format, v...)
		return
	}

	globalLogger.Errorf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	if globalLogger == nil {
		log.Printf(format, v...)
		return
	}

	globalLogger.Warnf(format, v...)
}

func Infof(format string, v ...interface{}) {
	if globalLogger == nil {
		log.Printf(format, v...)
		return
	}

	globalLogger.Infof(format, v...)
}

func Debugf(format string, v ...interface{}) {
	if globalLogger == nil {
		log.Printf(format, v...)
		return
	}

	globalLogger.Debugf(format, v...)
}
