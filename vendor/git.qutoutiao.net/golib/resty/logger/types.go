package logger

import "io"

// Interface is to abstract the logging from Resty. Gives control to
// the Resty users, choice of the logger.
type Interface interface {
	SetOutput(w io.Writer)
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}
