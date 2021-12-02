package handler

// handlerStore handler function map
var (
	handlerStore    = make(map[string]func() Handler)
	buildInHandlers = []string{BizkeeperConsumer, BizkeeperProvider, Loadbalance, Router, TracingConsumer, TracingProvider, RatelimiterConsumer, RatelimiterProvider, FaultInject}
)

// constant keys for handlers
const (
	//consumer chain
	Transport           = "transport"
	Loadbalance         = "loadbalance"
	BizkeeperConsumer   = "bizkeeper-consumer"
	TracingConsumer     = "tracing-consumer"
	RatelimiterConsumer = "ratelimiter-consumer"
	Router              = "router"
	FaultInject         = "fault-inject"

	//provider chain
	RatelimiterProvider = "ratelimiter-provider"
	TracingProvider     = "tracing-provider"
	BizkeeperProvider   = "bizkeeper-provider"
	MetricsProvider     = "metrics-provider"
	MetricsConsumer     = "metrics-consumer"
	LogProvider         = "log-provider"
)

// init is for to initialize the all handlers at boot time
func init() {
	//register build-in handler,don't need to call RegisterHandlerFunc
	handlerStore[Transport] = newTransportHandler
	handlerStore[Loadbalance] = newLBHandler
	handlerStore[BizkeeperConsumer] = newBizKeeperConsumerHandler
	handlerStore[BizkeeperProvider] = newBizKeeperProviderHandler
	handlerStore[RatelimiterConsumer] = newConsumerRateLimiterHandler
	handlerStore[RatelimiterProvider] = newProviderRateLimiterHandler
	handlerStore[TracingProvider] = newTracingProviderHandler
	handlerStore[TracingConsumer] = newTracingConsumerHandler
	handlerStore[FaultInject] = newFaultHandler
	handlerStore[MetricsProvider] = newMetricsProviderHandler
	handlerStore[MetricsConsumer] = newMetricsConsumerHandler
	handlerStore[LogProvider] = newLogProviderHandler
}
