package jaeger

import (
	"strconv"

	"git.qutoutiao.net/gopher/qms/internal/core/tracing"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qutracing"
	"github.com/opentracing/opentracing-go"
	jaegerstd "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
)

const (
	SamplingRateDef  = 1.0
	TraceFileNameDef = "/data/logs/trace/trace.log"
)

func init() {
	tracing.InstallTracer("jaeger", NewTracer)
}

func NewTracer(option *tracing.Option) (opentracing.Tracer, error) {
	var err error
	samplingRate := SamplingRateDef
	if option.SamplingRate != "" {
		samplingRate, err = strconv.ParseFloat(option.SamplingRate, 64)
		if err != nil {
			qlog.Errorf("parse sampling rate failed: %v", err)
			return nil, err
		}
	}

	var traceFileName = TraceFileNameDef
	if option.FileName != "" {
		traceFileName = option.FileName
	}

	var bufferSize int64
	if option.BufferSize > 0 {
		bufferSize = option.BufferSize
	}

	// 增加TextMap类型的Propagator
	options := make([]jaegerstd.TracerOption, 0)
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	options = append(options, jaegerstd.TracerOptions.Injector(opentracing.TextMap, zipkinPropagator))
	options = append(options, jaegerstd.TracerOptions.Extractor(opentracing.TextMap, zipkinPropagator))
	if option.MaxTagValueLength > 0 {
		options = append(options, jaegerstd.TracerOptions.MaxTagValueLength(option.MaxTagValueLength))
	}

	tracer, _, err := qutracing.NewJaegerTracer(runtime.ServiceName, traceFileName, samplingRate, nil, int(bufferSize), options...)
	if err != nil {
		return nil, err
	}

	qlog.Infof("init trace success: rate=%.4f, file=%s", samplingRate, traceFileName)
	return tracer, nil
}
