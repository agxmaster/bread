package log

import (
	"io"
	"log"
)

// 日志接口
type Logger interface {
	Error(msg string)
	Errorf(msg string, args ...interface{})

	Info(msg string)
	Infof(msg string, args ...interface{})

	io.Closer
}

var StdLogger = &stdLogger{}

type stdLogger struct{}

func (l *stdLogger) Error(msg string) {
	log.Printf("ERROR: %s", msg)
}

func (l *stdLogger) Errorf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (l *stdLogger) Close() error {
	return nil
}

func (l *stdLogger) Info(msg string) {
	log.Printf("INFO: %s", msg)
}
func (l *stdLogger) Infof(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

var NullLogger = &nullLogger{}

type nullLogger struct{}

func (l *nullLogger) Error(msg string)                       {}
func (l *nullLogger) Errorf(msg string, args ...interface{}) {}
func (l *nullLogger) Info(msg string)                        {}
func (l *nullLogger) Infof(msg string, args ...interface{})  {}
func (l *nullLogger) Close() error                           { return nil }
