package qudiscovery

import (
	"github.com/sirupsen/logrus"
)

var (
	// Log 全局日志对象
	Log = Logger(logrus.NewEntry(logrus.StandardLogger()))
)

// SetLogger 设置全局日志对象
func SetLogger(logger Logger) {
	Log = logger
}
