package logger

import (
	"io"
	"net/http"
	"sync"

	"git.qutoutiao.net/golib/resty/config"
	"github.com/rs/zerolog"
)

// RequestLog struct is used to collected information from resty request
// instance for debug logging. It sent to request log callback before resty
// actually logs the information.
type RequestLog struct {
	Header http.Header
	Body   string
}

// ResponseLog struct is used to collected information from resty response
// instance for debug logging. It sent to response log callback before resty
// actually logs the information.
type ResponseLog struct {
	Header http.Header
	Body   string
}

func New() Interface {
	return NewWithPrefix("")
}

func NewWithPrefix(module string) Interface {
	nlog := zlog.Hook(zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
		e.Str("resty", config.Version)

		if len(module) > 0 {
			e.Str("module", module)
		}

	})).With().CallerWithSkipFrameCount(3).Logger()

	return &logger{
		zlog: nlog,
	}
}

type logger struct {
	mux  sync.Mutex
	zlog zerolog.Logger
}

var _ Interface = (*logger)(nil)

func (l *logger) SetOutput(w io.Writer) {
	l.mux.Lock()
	if w != nil {
		l.zlog = l.zlog.Output(w)
	}
	l.mux.Unlock()
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.zlog.Error().Msgf(format, v...)
}

func (l *logger) Warnf(format string, v ...interface{}) {
	l.zlog.Warn().Msgf(format, v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.zlog.Debug().Msgf(format, v...)
}
