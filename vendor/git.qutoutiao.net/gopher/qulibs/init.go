package qulibs

var (
	// stdlog is use for qulibs components, such as Init() with error.
	stdlog = NewLoggerWithSkip(LogInfo, 4)
)

func Info(args ...interface{}) {
	stdlog.Info(args...)
}

func Infof(format string, args ...interface{}) {
	stdlog.Infof(format, args...)
}

func Warn(args ...interface{}) {
	stdlog.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	stdlog.Warnf(format, args)
}

func Error(args ...interface{}) {
	stdlog.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	stdlog.Errorf(format, args...)
}
