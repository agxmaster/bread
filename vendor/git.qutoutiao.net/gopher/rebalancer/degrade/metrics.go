package degrade

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// 服务降级 进入自我保护的时间
	EeventDegradeSelfProtectionTime = "eventDegradeSelfProtectionTime"
	// 服务降级 进入恐慌的状态
	EventDegradePanicStatus = "eventDegradePanicStatus"
	// 服务熔断 进入保护模式
	EventBreakerOpenStatus = "eventBreakerOpenStatus"
	// 节点熔断 半开状态
	EventBreakerHalfOpenStatus = "eventBreakerHalfOpenStatus"

	// 服务降级 进入恐慌的 metrics
	MetricsDegradePanicStatus = 1
)

var (
	events   *prometheus.GaugeVec
	hostname string
)

func init() {
	hostname, _ = os.Hostname()
	events = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "degrade_events",
	}, []string{"hostname", "cluster", "event"})
	prometheus.MustRegister(events)
}

// EventSet 事件值设置
func EventSet(event, cluster string, value float64) {
	events.WithLabelValues(hostname, cluster, event).Set(value)
}

// EventDeleteLabelValues 删除无用事件
func EventDeleteLabelValues(event, cluster string) {
	events.DeleteLabelValues(hostname, cluster, event)
}
