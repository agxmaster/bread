package qutracing

import (
	"io"
	"os"
	"time"

	"golang.org/x/exp/rand"

	"git.qutoutiao.net/gopher/qulibs/config/meta"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

// NewJaegerTracer构造返回一个opentracing.Tracer。采样率小于0会被修正为0; 修正后的采样率
// 大于1，使用限速采样，每秒最多采样指定数量的trace; 修正后的采样率介于0和1之间则使用概率采样。
func NewJaegerTracer(serviceName, traceFileName string, samplingRate float64, sig os.Signal, bufferSize int, options ...jaeger.TracerOption) (opentracing.Tracer, io.Closer, error) {
	if paasServiceName := meta.App(); paasServiceName != "" {
		serviceName = paasServiceName
	}

	var sampler jaeger.Sampler
	if samplingRate < 0 {
		samplingRate = 0
	}
	if samplingRate <= 1 {
		sampler, _ = NewProbabilisticSampler(samplingRate)
	} else {
		sampler = jaeger.NewRateLimitingSampler(samplingRate)
	}

	if traceFileName == "" {
		traceFileName = "/data/logs/trace.log"
	}

	reporter := NewJaegerFileReporter(traceFileName, JaegerReopenSignal(sig), JaegerBufferSize(bufferSize))

	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	injector := jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, zipkinPropagator)
	extractor := jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, zipkinPropagator)
	throtter := jaeger.TracerOptions.DebugThrottler(DefaultThrottler{})
	randomNumber := jaeger.TracerOptions.RandomNumber(defaultRandomNumber)

	options = append(options, injector, extractor, throtter, randomNumber)

	// create Jaeger tracer
	tracer, closer := jaeger.NewTracer(
		serviceName,
		sampler,
		reporter,
		options...,
	)

	//opentracing.SetGlobalTracer(tracer)
	return tracer, closer, nil
}

// randomNumber 冲突解决随机算法
// 0x1000000000000000 保证转化成16进制没有前缀0
func defaultRandomNumber() uint64 {
	return uint64(rand.Int63() | 0x1000000000000000)
}

// DefaultThrottler doesn't throttle at all.
type DefaultThrottler struct{}

// IsAllowed implements Throttler#IsAllowed.
func (t DefaultThrottler) IsAllowed(operation string) bool {
	return true
}

func buildSampler(samplingRate float64) jaeger.Sampler {
	if samplingRate > 1.0 {
		return jaeger.NewRateLimitingSampler(samplingRate)
	}
	if samplingRate < 0 {
		samplingRate = 0
	}
	sampler, _ := jaeger.NewProbabilisticSampler(samplingRate)
	return sampler
}
