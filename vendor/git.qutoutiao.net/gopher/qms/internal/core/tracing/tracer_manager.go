package tracing

import (
	"fmt"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"github.com/opentracing/opentracing-go"
)

// TracerFuncMap saves NewTracer func
// key: impl name
// val: tracer new func
var TracerFuncMap = make(map[string]NewTracer)

// NewTracer is the func to return global tracer
type NewTracer func(o *Option) (opentracing.Tracer, error)

//InstallTracer install new opentracing tracer
func InstallTracer(name string, f NewTracer) {
	TracerFuncMap[name] = f
	qlog.Trace("Installed tracing plugin: " + name)

}

// GetTracerFunc get NewTracer
func GetTracerFunc(name string) (NewTracer, error) {
	tracer, ok := TracerFuncMap[name]
	if !ok {
		return nil, fmt.Errorf("not supported tracer [%s]", name)
	}
	return tracer, nil
}

// Init initialize the global tracer
func Init() error {
	qlog.Trace("Tracing enabled. Start to init tracer.")
	f, err := GetTracerFunc("jaeger")
	if err != nil {
		qlog.Warn("can not load any opentracing plugin, lost distributed tracing function")
		return nil
	}
	trace := config.Get().Trace.Setting
	tracer, err := f(&Option{
		SamplingRate:      trace.SamplingRate,
		MaxTagValueLength: trace.MaxTagValueLength,
		FileName:          trace.TraceFile,
	})
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
