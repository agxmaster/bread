package archaius

import (
	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/config/model"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/qpslimiter"
	"git.qutoutiao.net/gopher/qms/third_party/forked/afex/hystrix-go/hystrix"
)

//Panel pull configs from archaius
type Panel struct {
}

func newPanel(options control.Options) control.Panel {
	SaveToLBCache(config.GetLoadBalancing())
	SaveToCBCache(config.GetHystrixConfig())
	return &Panel{}
}

//GetCircuitBreaker return command , and circuit breaker settings
func (p *Panel) GetCircuitBreaker(inv *invocation.Invocation, serviceType string) (string, hystrix.CommandConfig) {
	key := GetCBCacheKey(inv.MicroServiceName, serviceType)
	command := control.NewCircuitName(serviceType, coreconf.GetHystrixConfig().CircuitBreakerProperties.Scope, inv)
	c, ok := CBConfigCache.Get(key)
	if ok {
		return command, c.(hystrix.CommandConfig)
	}

	configed := false // 暂时不支持provider侧熔断 configed := archaius.GetBool(coreconf.GetCircuitBreakerEnabledKey(fmt.Sprintf("%s.%s", serviceType, inv.MicroServiceName)), false)
	if serviceType == common.Consumer {
		configed = config.GetUpstream(inv.MicroServiceName).CircuitBreaker.Enabled
	}
	if configed {
		saveEachCB(inv.MicroServiceName, serviceType)
		c, ok = CBConfigCache.Get(key)
		if ok {
			return command, c.(hystrix.CommandConfig)
		}
	}

	c, _ = CBConfigCache.Get(serviceType)
	return command, c.(hystrix.CommandConfig)
}

//GetLoadBalancing get load balancing config
func (p *Panel) GetLoadBalancing(inv *invocation.Invocation) control.LoadBalancingConfig {
	c, ok := LBConfigCache.Get(inv.MicroServiceName)
	if !ok {
		c, ok = LBConfigCache.Get("")
		if !ok {
			return DefaultLB

		}
		return c.(control.LoadBalancingConfig)

	}
	return c.(control.LoadBalancingConfig)

}

//GetRateLimiting get rate limiting config
func (p *Panel) GetRateLimiting(inv *invocation.Invocation, serviceType string) (rl control.RateLimitingConfig) {
	if serviceType == common.Consumer {
		if !config.GetUpstream(inv.MicroServiceName).RateLimit.Enabled {
			return
		}
		keys := qpslimiter.GetConsumerKey(inv.MicroServiceName, inv.OperationID)
		rl.Rate, rl.Key = qpslimiter.GetQPSTrafficLimiter().GetUpstreamQPSRate(keys)
	} else {
		// TODO：以前provider是全局，现在是针对单个服务，要匹配好服务名
		if !config.GetService(inv.MicroServiceName).RateLimit.Enabled {
			return
		}
		keys := qpslimiter.GetProviderKey(inv.MicroServiceName, inv.OperationID)
		rl.Rate, rl.Key = qpslimiter.GetQPSTrafficLimiter().GetServiceQPSRate(keys)
	}

	if rl.Rate == qpslimiter.DefaultRate {
		return
	}
	rl.Enabled = true

	return
}

//GetFaultInjection get Fault injection config
func (p *Panel) GetFaultInjection(inv *invocation.Invocation) model.Fault {
	return model.Fault{}

}

//GetEgressRule get egress config
func (p *Panel) GetEgressRule() []control.EgressConfig {
	return []control.EgressConfig{}
}

func init() {
	control.InstallPlugin("archaius", newPanel)
}
