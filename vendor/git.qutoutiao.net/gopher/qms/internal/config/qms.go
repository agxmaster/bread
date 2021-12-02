package config

import (
	"sync"

	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/config/model"
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

// 原则：除了initiator需要提前用到的可以不定义 不然都需要定义

type Config struct {
	Qms Qms `yaml:"qms"`
}

// Qms 相关配置
type Qms struct {
	upstreamAlias map[string]string   // 别名就是address名称 兼容用
	Upstreams     map[string]Upstream `yaml:"upstreams"`
	Service       Service
	Registry      Registry
	Metrics       Metrics
	Healthy       Healthy
	Trace         Trace
	AccessLog     AccessLog
	PProf         PProf
	Native        Native
}

var (
	conf     Config
	confOnce sync.Once
)

// Get 获取qms配置
func Get() *Qms {
	return &conf.Qms
}

func Init() (err error) {
	confOnce.Do(func() {
		qconf.Reset() // 重置配置
		qconf.AddOptionalFile(fileutil.AppConfigPath(), fileutil.AdvancedConfigPath(), fileutil.UpstreamConfigPath())
		if err = qconf.ReadInConfig(); err != nil { // read in memory
			return
		}

		// 初始化upstream
		if err = initUpstream(); err != nil { // 依赖archaius先初始化
			qlog.Errorf("解析upstream配置文件失败: %v", err)
			return
		}

		// 初始化service
		initService()

		// 初始化服务治理
		initAdvance()

		if err = coreconf.Init(); err != nil {
			return
		}

		service := Get().Service
		if err = runtime.Init(&runtime.Service{
			Service:     service.AppID,
			Environment: service.Env,
			Version:     service.Version,
		}); err != nil {
			return
		}
	})

	return
}

// GetHystrixConfig return cb config
func GetHystrixConfig() *model.HystrixConfig {
	hystrixCnf := coreconf.GetHystrixConfig()

	// 替换upstream配置
	commonUpstream := GetUpstream(constutil.Common)
	transport := hystrixCnf.IsolationProperties.Consumer
	transport.MaxConcurrentRequests = commonUpstream.Transport.MaxConcurrent
	transport.TimeoutInMilliseconds = commonUpstream.Transport.TimeoutMs
	transport.AnyService = make(map[string]model.IsolationSpec)

	circuit := hystrixCnf.CircuitBreakerProperties.Consumer
	circuit.Enabled = commonUpstream.CircuitBreaker.Enabled
	circuit.ForceOpen = commonUpstream.CircuitBreaker.ForceOpen
	circuit.ForceClose = commonUpstream.CircuitBreaker.ForceClosed
	circuit.SleepWindowInMilliseconds = commonUpstream.CircuitBreaker.SleepWindowMs
	circuit.RequestVolumeThreshold = commonUpstream.CircuitBreaker.RequestVolumeThreshold
	circuit.ErrorThresholdPercentage = commonUpstream.CircuitBreaker.ErrorThresholdPercentage
	circuit.AnyService = make(map[string]model.CircuitBreakPropertyStruct)
	hystrixCnf.CircuitBreakerProperties.Scope = commonUpstream.CircuitBreaker.Scope

	for key, value := range conf.Qms.Upstreams {
		if key == constutil.Common {
			continue
		}
		transport.AnyService[key] = model.IsolationSpec{
			TimeoutInMilliseconds: value.Transport.TimeoutMs,
			MaxConcurrentRequests: value.Transport.MaxConcurrent,
		}

		circuit.AnyService[key] = model.CircuitBreakPropertyStruct{
			Enabled:                   value.CircuitBreaker.Enabled,
			ForceOpen:                 value.CircuitBreaker.ForceOpen,
			ForceClose:                value.CircuitBreaker.ForceClosed,
			SleepWindowInMilliseconds: value.CircuitBreaker.SleepWindowMs,
			RequestVolumeThreshold:    value.CircuitBreaker.RequestVolumeThreshold,
			ErrorThresholdPercentage:  value.CircuitBreaker.ErrorThresholdPercentage,
		}
	}

	return hystrixCnf
}

//GetLoadBalancing return lb config
func GetLoadBalancing() *model.LoadBalancing {
	upstream := GetUpstream(constutil.Common)
	loadbalance := &model.LoadBalancing{
		Enabled: true,
		Strategy: map[string]string{
			"name": upstream.Transport.Loadbalance,
		},
		//RetryEnabled:   upstream.Retry.Enabled,
		//RetryOnNext:    upstream.Retry.OnNext,
		//RetryOnSame:    upstream.Retry.OnSame,
		//RetryCondition: upstream.Retry.Condition,
		//Backoff: model.BackoffStrategy{
		//	Kind:  upstream.Retry.Backoff.Kind,
		//	MinMs: upstream.Retry.Backoff.MinMs,
		//	MaxMs: upstream.Retry.Backoff.MaxMs,
		//},
		AnyService: make(map[string]model.LoadBalancingSpec),
	}

	for key, value := range conf.Qms.Upstreams {
		if key == constutil.Common {
			continue
		}

		loadbalance.AnyService[key] = model.LoadBalancingSpec{
			Strategy: map[string]string{
				"name": value.Transport.Loadbalance,
			},
			//RetryEnabled: value.Retry.Enabled,
			//RetryOnNext:  value.Retry.OnNext,
			//RetryOnSame:  value.Retry.OnSame,
			//Backoff: model.BackoffStrategy{
			//	Kind:  value.Retry.Backoff.Kind,
			//	MinMs: value.Retry.Backoff.MinMs,
			//	MaxMs: value.Retry.Backoff.MaxMs,
			//},
		}
	}

	return loadbalance
}

// GetContractDiscoveryDisable returns the Disable of contract discovery registry
func GetContractDiscoveryDisable() bool {
	if b := qconf.GetBool("qms.service.registry.contractDiscovery.disabled", false); b {
		return b
	}
	return Get().Registry.Disabled
}
