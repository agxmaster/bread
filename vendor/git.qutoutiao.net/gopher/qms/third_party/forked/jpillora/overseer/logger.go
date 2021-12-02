package overseer

type logger interface {
	Tracef(format string, args ...interface{}) // 格式化并记录 TraceLevel 级别的日志
	Debugf(format string, args ...interface{}) // 格式化并记录 DebugLevel 级别的日志
	Infof(format string, args ...interface{})  // 格式化并记录 InfoLevel 级别的日志
	Warnf(format string, args ...interface{})  // 格式化并记录 WarnLevel 级别的日志
	Errorf(format string, args ...interface{}) // 格式化并记录 ErrorLevel 级别的日志
}

type wrapLogger struct {
	prefix string
	logger
}

func newWrapLogger(prefix string, logger logger) *wrapLogger {
	return &wrapLogger{
		prefix: prefix,
		logger: logger,
	}
}

func (w *wrapLogger) Tracef(format string, args ...interface{}) {
	w.logger.Tracef(w.prefix+format, args...)
}

func (w *wrapLogger) Debugf(format string, args ...interface{}) {
	w.logger.Debugf(w.prefix+format, args...)
}

func (w *wrapLogger) Infof(format string, args ...interface{}) {
	w.logger.Infof(w.prefix+format, args...)
}

func (w *wrapLogger) Warnf(format string, args ...interface{}) {
	w.logger.Warnf(w.prefix+format, args...)
}

func (w *wrapLogger) Errorf(format string, args ...interface{}) {
	w.logger.Errorf(w.prefix+format, args...)
}
