package logger

import (
	"io"

	"git.qutoutiao.net/pedestal/discovery/logger/hclog"
	stdhclog "github.com/hashicorp/go-hclog"
)

type Logger struct {
	hlog stdhclog.Logger
}

func New(w ...io.Writer) *Logger {
	return NewWithHclog(hclog.NewWithSkipFrameCount(5, w...))
}

func NewWithHclog(hlog stdhclog.Logger) *Logger {
	return &Logger{hlog: hlog}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.hlog.Error(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.hlog.Warn(format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.hlog.Info(format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.hlog.Debug(format, v...)
}
