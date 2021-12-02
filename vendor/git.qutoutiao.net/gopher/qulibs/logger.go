package qulibs

import (
	"github.com/dolab/logger"
)

const (
	// LogOff states that no logging should be performed by the SDK. This is the
	// default state of the SDK, and should be use to disable all logging.
	LogOff LogLevelType = iota * 0x1000

	// LogDebug state that debug output should be logged by the SDK. This should
	// be used to inspect request made and responses received.
	LogDebug
)

// log levels
const (
	LogInfo = LogDebug | (1 << iota)
	LogWarn
	LogError
)

// NewLogger returns a Logger with level given.
func NewLogger(level LogLevelType) Logger {
	return NewLoggerWithSkip(level, 3)
}

func NewLoggerWithSkip(level LogLevelType, skip int) Logger {
	log, _ := logger.New("stderr")

	if skip > 3 {
		log.SetSkip(skip)
	} else {
		log.SetSkip(3)
	}

	return &dummyLogger{
		level: level,
		log:   log,
	}
}

// A dummyLogger implements qulibs Logger interface.
type dummyLogger struct {
	level LogLevelType
	log   *logger.Logger
}

// NewDummyLogger returns a Logger with LogOff level, which means no logs will be wrote.
func NewDummyLogger() Logger {
	return NewLogger(LogOff)
}

// Debug of dummy logger
func (l *dummyLogger) Debug(args ...interface{}) {
	if l.level.AtLeast(LogDebug) {
		return
	}

	l.log.Debug(args...)
}

// Debugf of dummy logger
func (l *dummyLogger) Debugf(format string, args ...interface{}) {
	if l.level.AtLeast(LogDebug) {
		return
	}
	if format[len(format)-1] != '\n' {
		format += "\n"
	}

	l.log.Debugf(format, args...)
}

// Info of dummy logger
func (l *dummyLogger) Info(args ...interface{}) {
	if l.level.AtLeast(LogInfo) {
		return
	}

	l.log.Info(args...)
}

// Infof of dummy logger
func (l *dummyLogger) Infof(format string, args ...interface{}) {
	if l.level.AtLeast(LogInfo) {
		return
	}

	l.log.Infof(format, args...)
}

// Warn of dummy logger
func (l *dummyLogger) Warn(args ...interface{}) {
	if l.level.AtLeast(LogWarn) {
		return
	}

	l.log.Warn(args...)
}

// Warnf of dummy logger
func (l *dummyLogger) Warnf(format string, args ...interface{}) {
	if l.level.AtLeast(LogWarn) {
		return
	}

	l.log.Warnf(format, args...)
}

// Error of dummy logger
func (l *dummyLogger) Error(args ...interface{}) {
	if l.level.AtLeast(LogError) {
		return
	}

	l.log.Error(args...)
}

// Errorf of dummy logger
func (l *dummyLogger) Errorf(format string, args ...interface{}) {
	if l.level.AtLeast(LogError) {
		return
	}

	l.log.Errorf(format, args...)
}
