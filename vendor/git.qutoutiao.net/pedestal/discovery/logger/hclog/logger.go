package hclog

import (
	"io"
	stdlog "log"
	"os"

	"git.qutoutiao.net/pedestal/discovery/metrics"
	"github.com/golib/zerolog"
	"github.com/hashicorp/go-hclog"
)

var (
	zlog = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

type hclogger struct {
	zlog zerolog.Logger
	args []interface{}
}

func New(w ...io.Writer) hclog.Logger {
	return NewWithSkipFrameCount(5, w...)
}

func NewWithSkipFrameCount(skip int, w ...io.Writer) hclog.Logger {
	if skip <= 0 {
		skip = 3
	}

	nlog := zlog.Hook(zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
		e.Str("discovery", metrics.Version)
	})).With().CallerWithSkipFrameCount(skip).Logger()
	if len(w) > 0 {
		nlog = nlog.Output(w[0])
	}

	return &hclogger{
		zlog: nlog,
	}
}

func (log *hclogger) With(args ...interface{}) hclog.Logger {
	log.args = args

	return log
}

func (log *hclogger) Trace(msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) Log(level hclog.Level, msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) Debug(msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) Info(msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) Warn(msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) Error(msg string, args ...interface{}) {
	if len(log.args) > 0 {
		args = append(log.args, args...)
	}

	log.zlog.Printf(msg, args...)
}

func (log *hclogger) IsTrace() bool {
	level := log.zlog.GetLevel()

	return hclog.LevelFromString(level.String()) >= hclog.Trace
}

func (log *hclogger) IsDebug() bool {
	level := log.zlog.GetLevel()

	return hclog.LevelFromString(level.String()) >= hclog.Debug
}

func (log *hclogger) IsInfo() bool {
	level := log.zlog.GetLevel()

	return hclog.LevelFromString(level.String()) >= hclog.Info
}

func (log *hclogger) IsWarn() bool {
	level := log.zlog.GetLevel()

	return hclog.LevelFromString(level.String()) >= hclog.Warn
}

func (log *hclogger) IsError() bool {
	level := log.zlog.GetLevel()

	return hclog.LevelFromString(level.String()) >= hclog.Error
}

func (log *hclogger) ImpliedArgs() []interface{} {
	return log.args
}

func (log *hclogger) Name() string {
	return "discovery"
}

func (log *hclogger) Named(name string) hclog.Logger {
	nlog := zlog.Hook(zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
		e.Strs("sdk", []string{"discovery", name})
	})).With().Logger()

	return &hclogger{zlog: nlog, args: log.args}
}

func (log *hclogger) ResetNamed(name string) hclog.Logger {
	return log.Named(name)
}

func (log *hclogger) SetLevel(level hclog.Level) {
	switch zerolog.Level(level) {
	case zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel, zerolog.NoLevel, zerolog.Disabled, zerolog.TraceLevel:
		log.zlog = log.zlog.Level(zerolog.Level(level))
	}
}

func (log *hclogger) StandardLogger(opts *hclog.StandardLoggerOptions) *stdlog.Logger {
	zlevel := zerolog.InfoLevel
	switch zerolog.Level(opts.ForceLevel) {
	case zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel, zerolog.NoLevel, zerolog.Disabled, zerolog.TraceLevel:
		zlevel = zerolog.Level(opts.ForceLevel)
	}

	return stdlog.New(zlog.Level(zlevel), "["+log.zlog.GetLevel().String()+"]", stdlog.LstdFlags|stdlog.Lshortfile)
}

func (log *hclogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	zlevel := zerolog.InfoLevel
	switch zerolog.Level(opts.ForceLevel) {
	case zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel, zerolog.NoLevel, zerolog.Disabled, zerolog.TraceLevel:
		zlevel = zerolog.Level(opts.ForceLevel)
	}

	return zlog.Level(zlevel)
}
