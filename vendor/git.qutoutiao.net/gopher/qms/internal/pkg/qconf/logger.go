package qconf

type logger interface {
	Tracef(format string, args ...interface{}) // 格式化并记录 TraceLevel 级别的日志
	Debugf(format string, args ...interface{}) // 格式化并记录 DebugLevel 级别的日志
	Infof(format string, args ...interface{})  // 格式化并记录 InfoLevel 级别的日志
	Warnf(format string, args ...interface{})  // 格式化并记录 WarnLevel 级别的日志
	Errorf(format string, args ...interface{}) // 格式化并记录 ErrorLevel 级别的日志

	Printf(format string, args ...interface{})
	Println(format string, args ...interface{})
	Fatal(args ...interface{})
}

//type wrapLogger struct {
//	logger
//}

//func newWrapLogger(logger logger) *wrapLogger {
//	return &wrapLogger{
//		logger: logger,
//	}
//}

//func (w *wrapLogger) Tracef(format string, args ...interface{}) {
//	w.logger.Tracef(format, args...)
//}
//
//func (w *wrapLogger) Debugf(format string, args ...interface{}) {
//	w.logger.Debugf(format, args...)
//}
//
//func (w *wrapLogger) Infof(format string, args ...interface{}) {
//	w.logger.Infof(format, args...)
//}
//
//func (w *wrapLogger) Warnf(format string, args ...interface{}) {
//	w.logger.Warnf(format, args...)
//}
//
//func (w *wrapLogger) Errorf(format string, args ...interface{}) {
//	w.logger.Errorf(format, args...)
//}
//
//func (w *wrapLogger) Printf(format string, args ...interface{}) {
//	w.logger.Printf(format, args...)
//}
//
//func (w *wrapLogger) Println(format string, args ...interface{}) {
//	w.logger.Println(format, args...)
//}
//
//func (w *wrapLogger) Fatal(args ...interface{}) {
//	w.logger.Fatal(args...)
//}
