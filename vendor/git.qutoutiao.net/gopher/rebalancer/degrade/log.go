package degrade

import (
	"io"

	"github.com/sirupsen/logrus"
)

// Logger 日志
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Writer() *io.PipeWriter
}

var (
	// Log 全局日志对象
	Log = Logger(logrus.NewEntry(logrus.StandardLogger()))
)

// SetLogger 设置全局日志对象
func SetLogger(logger Logger) {
	Log = logger
}
