package config

import (
	"strings"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type Upstream struct {
	Address    string     `yaml:"address"`    // PassID、域名、IP:Port
	Sidecar    Sidecar    `yaml:"sidecar"`    // sidecar
	Transport  Transport  `yaml:"transport"`  // 传输
	Discoverer Discoverer `yaml:"discoverer"` // 服务发现
	//Retry          Retry          `yaml:"retry"`             // 重试
	RateLimit      RateLimit      `yaml:"rate_limit"`        // 限流
	CircuitBreaker CircuitBreaker `yaml:"circuit_breaker"`   // 熔断
	Parampath      []string       `yaml:"parampath,flow"`    // 指定参数路由 防止metrics过度膨胀
	CustomRoute    []CustomRoute  `yaml:"custom_route,flow"` // 自定义服务路由
	LoadBalance    LoadBalance    `yaml:"load_balance"`
	Env            qenv.Env
}

type Transport struct {
	TimeoutMs     int    `yaml:"timeout_ms"`      // 超时时间
	Loadbalance   string `yaml:"loadbalance"`     // 负载均衡
	MaxConcurrent int    `yaml:"max_concurrent"`  // 最大并发数
	RespCacheMs   int    `yaml:"resp_cache_ms"`   // 返回[GET]缓存时间
	RespCacheSize int    `yaml:"resp_cache_size"` // 返回[GET]缓存大小
	MaxIdleConn   int    `yaml:"max_idle_conn"`   // 最大空闲连接数
	MaxBodyBytes  int    `yaml:"max_body_bytes"`  // 最大传输的body大小
}

type Discoverer struct {
	Pilot      string   `yaml:"pilot"` // 统一发现服务
	Datacenter string   `yaml:"dc"`    // 数据中心
	Tags       []string `yaml:"tags"`  // 需要过滤的tag
}

type Retry struct {
	Enabled   bool    `yaml:"enabled"`
	OnNext    int     `yaml:"on_next"`
	OnSame    int     `yaml:"on_same"`
	Condition string  `yaml:"condition"`
	Backoff   Backoff `yaml:"backoff"`
}

type Backoff struct {
	Kind  string `yaml:"kind"`
	MinMs int    `yaml:"min_ms"`
	MaxMs int    `yaml:"max_ms"`
}

type RateLimit struct {
	Enabled bool           `yaml:"enabled"`
	Limit   map[string]int `yaml:",inline"`
}

type CircuitBreaker struct {
	Enabled                  bool   `yaml:"enabled"`
	Scope                    string `yaml:"scope"`
	ForceOpen                bool   `yaml:"force_open"`
	ForceClosed              bool   `yaml:"force_closed"`
	SleepWindowMs            int    `yaml:"sleep_window_ms"`
	RequestVolumeThreshold   int    `yaml:"request_volume_threshold"`
	ErrorThresholdPercentage int    `yaml:"error_threshold_percentage"`
}

type CustomRoute struct {
	Address string `yaml:"address"`
	Weight  int    `yaml:"weight"`
}

type LoadBalance struct {
	Strategy       string `yaml:"strategy"`        // 策略
	WithoutBreaker bool   `yaml:"without_breaker"` //不包含熔断
}

type Sidecar struct {
	Enabled     bool   `yaml:"enabled"`
	Address     string `yaml:"address"`
	MeshService string `yaml:"mesh_service"`
}

func (up Upstream) copy() Upstream {
	upstream := up

	upstream.Discoverer.Tags = make([]string, len(up.Discoverer.Tags))
	copy(upstream.Discoverer.Tags, up.Discoverer.Tags)

	upstream.Parampath = make([]string, len(up.Parampath))
	copy(upstream.Parampath, up.Parampath)

	upstream.CustomRoute = make([]CustomRoute, len(up.CustomRoute))
	copy(upstream.CustomRoute, up.CustomRoute)

	limit := make(map[string]int)
	for key, value := range up.RateLimit.Limit {
		limit[key] = value
	}
	upstream.RateLimit.Limit = limit

	return upstream
}

func GetUpstream(name string) *Upstream {
	upstreams := Get().Upstreams

	name = getRealName(name)
	value, ok := upstreams[name]
	if !ok {
		value = upstreams[constutil.Common]
	}

	return &value
}

func GetTimeoutDuration(service string) time.Duration {
	return time.Duration(GetUpstream(service).Transport.TimeoutMs) * time.Millisecond
}

func GetRespCacheDuration(service string) time.Duration {
	return time.Duration(GetUpstream(service).Transport.RespCacheMs) * time.Millisecond
}

func GetRespCacheSize(service string) int {
	return GetUpstream(service).Transport.RespCacheSize
}

func GetRemoteService(service string) string {
	if endpoint := GetUpstream(service).Address; endpoint != "" {
		return endpoint
	}
	return service
}

/*
 * 先设置COMMON配置
 * 再设置Service配置
 * 最后兼容老的且不匹配Service的配置
 */
func initUpstream() (err error) {
	if err = unmarshalAdvance(); err != nil { // 先加载advance.yaml为了兼容
		return errors.WithStack(err)
	}

	conf.Qms.upstreamAlias = make(map[string]string)
	conf.Qms.Upstreams = make(map[string]Upstream)
	conf.Qms.Upstreams[constutil.Common] = newUpstream(constutil.Common)

	for service := range qconf.GetStringMap("qms.upstreams") {
		if service == constutil.Common {
			continue
		}
		upstream := newUpstream(service)
		conf.Qms.upstreamAlias[upstream.Address] = service
		conf.Qms.Upstreams[service] = upstream
	}

	// 兼容老的配置(要把老的配置转成新的)
	inheritMessy()

	return
}

func newUpstream(service string) Upstream {
	// 继承default
	upstream := defaultUpstream(service)
	upstream.Address = qconf.GetString(getUpstreamKey(service, "address"), upstream.Address)
	upstream.Sidecar = Sidecar{
		Enabled:     qconf.GetBool(getUpstreamKey(service, "sidecar.enabled"), qconf.GetBool(getUpstreamKey(service, "sidecar_enabled"), upstream.Sidecar.Enabled)),
		Address:     qconf.GetString(getUpstreamKey(service, "sidecar.address"), upstream.Sidecar.Address),
		MeshService: qconf.GetString(getUpstreamKey(service, "sidecar.mesh_service"), upstream.Sidecar.MeshService),
	}
	upstream.Transport = Transport{
		TimeoutMs:     qconf.GetInt(getUpstreamKey(service, "transport.timeout_ms"), qconf.GetInt(getKey("qms.isolation.Consumer", service, "timeoutInMilliseconds"), upstream.Transport.TimeoutMs)),
		Loadbalance:   qconf.GetString(getUpstreamKey(service, "transport.loadbalance"), upstream.Transport.Loadbalance),
		RespCacheMs:   qconf.GetInt(getUpstreamKey(service, "transport.resp_cache_ms"), qconf.GetInt(getKey("qms.isolation.Consumer", service, "respCacheMs"), upstream.Transport.RespCacheMs)),
		RespCacheSize: qconf.GetInt(getUpstreamKey(service, "transport.resp_cache_size"), upstream.Transport.RespCacheSize),
		MaxConcurrent: qconf.GetInt(getUpstreamKey(service, "transport.max_concurrent"), qconf.GetInt(getKey("qms.isolation.Consumer", service, "maxConcurrentRequests"), upstream.Transport.MaxConcurrent)),
		MaxIdleConn:   qconf.GetInt(getUpstreamKey(service, "transport.max_idle_conn"), upstream.Transport.MaxIdleConn),
		MaxBodyBytes:  qconf.GetInt(getUpstreamKey(service, "transport.max_body_bytes"), upstream.Transport.MaxBodyBytes),
	}
	upstream.Discoverer = Discoverer{
		Datacenter: qconf.GetString(getUpstreamKey(service, "discoverer.dc"), upstream.Discoverer.Datacenter),
		Tags:       qconf.GetStringSlice(getUpstreamKey(service, "discoverer.tags"), upstream.Discoverer.Tags),
	}
	//upstream.Retry = Retry{
	//	Enabled:   qconf.GetBool(getUpstreamKey(service, "retry.enabled"), upstream.Retry.Enabled),
	//	OnNext:    qconf.GetInt(getUpstreamKey(service, "retry.on_next"), upstream.Retry.OnNext),
	//	OnSame:    qconf.GetInt(getUpstreamKey(service, "retry.on_same"), upstream.Retry.OnSame),
	//	Condition: qconf.GetString(getUpstreamKey(service, "retry.condition"), upstream.Retry.Condition),
	//	Backoff: Backoff{
	//		Kind:  qconf.GetString(getUpstreamKey(service, "retry.backoff.kind"), upstream.Retry.Backoff.Kind),
	//		MinMs: qconf.GetInt(getUpstreamKey(service, "retry.backoff.min_ms"), upstream.Retry.Backoff.MinMs),
	//		MaxMs: qconf.GetInt(getUpstreamKey(service, "retry.backoff.max_ms"), upstream.Retry.Backoff.MaxMs),
	//	},
	//}
	upstream.RateLimit = RateLimit{
		Enabled: qconf.GetBool(getUpstreamKey(service, "rate_limit.enabled"), upstream.RateLimit.Enabled),
		Limit: map[string]int{
			Limit: qconf.GetInt(getUpstreamKey(service, "rate_limit.limit"), upstream.RateLimit.Limit[Limit]),
		},
	}
	upstream.CircuitBreaker = CircuitBreaker{
		Enabled:                  qconf.GetBool(getUpstreamKey(service, "circuit_breaker.enabled"), qconf.GetBool(getKey("qms.circuitBreaker.Consumer", service, "enabled"), upstream.CircuitBreaker.Enabled)),
		ForceOpen:                qconf.GetBool(getUpstreamKey(service, "circuit_breaker.force_open"), qconf.GetBool(getKey("qms.circuitBreaker.Consumer", service, "forceOpen"), upstream.CircuitBreaker.ForceOpen)),
		ForceClosed:              qconf.GetBool(getUpstreamKey(service, "circuit_breaker.force_closed"), qconf.GetBool(getKey("qms.circuitBreaker.Consumer", service, "forceClosed"), upstream.CircuitBreaker.ForceClosed)),
		SleepWindowMs:            qconf.GetInt(getUpstreamKey(service, "circuit_breaker.sleep_window_ms"), qconf.GetInt(getKey("qms.circuitBreaker.Consumer", service, "sleepWindowInMilliseconds"), upstream.CircuitBreaker.SleepWindowMs)),
		RequestVolumeThreshold:   qconf.GetInt(getUpstreamKey(service, "circuit_breaker.request_volume_threshold"), qconf.GetInt(getKey("qms.circuitBreaker.Consumer", service, "requestVolumeThreshold"), upstream.CircuitBreaker.RequestVolumeThreshold)),
		ErrorThresholdPercentage: qconf.GetInt(getUpstreamKey(service, "circuit_breaker.error_threshold_percentage"), qconf.GetInt(getKey("qms.circuitBreaker.Consumer", service, "errorThresholdPercentage"), upstream.CircuitBreaker.ErrorThresholdPercentage)),
	}

	// 匹配ENV 如果没有就使用当前环境
	for _, tag := range upstream.Discoverer.Tags {
		if env := qenv.ToEnv(tag); env.IsValid() {
			upstream.Env = env
			break
		}
	}

	if service == constutil.Common {
		upstream.Sidecar.Enabled = qconf.GetBool(getUpstreamKey(service, "sidecar.enabled"), qconf.GetBool(getUpstreamKey(service, "sidecar_enabled"), qconf.GetBool("qms.sidecar.enabled", upstream.Sidecar.Enabled)))
		upstream.Transport.Loadbalance = qconf.GetString(getUpstreamKey(service, "transport.loadbalance"), qconf.GetString("qms.loadbalance.strategy.name", upstream.Transport.Loadbalance))
		upstream.RateLimit.Enabled = qconf.GetBool(getUpstreamKey(service, "rate_limit.enabled"), qconf.GetBool("qms.flowcontrol.Consumer.qps.enabled", upstream.RateLimit.Enabled))
		//upstream.Retry = Retry{
		//	Enabled:   qconf.GetBool(getUpstreamKey(service, "retry.enabled"), qconf.GetBool("qms.loadbalance.retryEnabled", upstream.Retry.Enabled)),
		//	OnNext:    qconf.GetInt(getUpstreamKey(service, "retry.on_next"), qconf.GetInt("qms.loadbalance.retryOnNext", upstream.Retry.OnNext)),
		//	OnSame:    qconf.GetInt(getUpstreamKey(service, "retry.on_same"), qconf.GetInt("qms.loadbalance.retryOnSame", upstream.Retry.OnSame)),
		//	Condition: qconf.GetString(getUpstreamKey(service, "retry.condition"), qconf.GetString("qms.loadbalance.retryCondition", upstream.Retry.Condition)),
		//	Backoff: Backoff{
		//		Kind:  qconf.GetString(getUpstreamKey(service, "retry.backoff.kind"), qconf.GetString("qms.loadbalance.backoff.kind", upstream.Retry.Backoff.Kind)),
		//		MinMs: qconf.GetInt(getUpstreamKey(service, "retry.backoff.min_ms"), qconf.GetInt("qms.loadbalance.backoff.MinMs", upstream.Retry.Backoff.MinMs)),
		//		MaxMs: qconf.GetInt(getUpstreamKey(service, "retry.backoff.max_ms"), qconf.GetInt("qms.loadbalance.backoff.MaxMs", upstream.Retry.Backoff.MaxMs)),
		//	},
		//}
		upstream.LoadBalance = LoadBalance{
			Strategy:       qconf.GetString(getUpstreamKey(service, "load_balance.strategy"), upstream.LoadBalance.Strategy),
			WithoutBreaker: qconf.GetBool(getUpstreamKey(service, "load_balance.without_breaker"), upstream.LoadBalance.WithoutBreaker),
		}
		return upstream
	}

	// 参数路由 COMMON没有
	for _, path := range qconf.GetStringSlice(getUpstreamKey(service, "parampath")) {
		upstream.Parampath = append(upstream.Parampath, path)
	}
	if len(upstream.Parampath) == 0 { // 兼容老的 不会合并
		for _, path := range qconf.GetStringSlice("qms.metrics.Consumer.parampath") {
			list := strings.SplitN(path, "/", 2)
			if len(list) < 2 {
				continue
			}
			if list[0] == service {
				upstream.Parampath = append(upstream.Parampath, "/"+list[1])
			}
		}
	}

	// 自定义服务路由
	routes := make([]interface{}, 0)
	for _, route := range qconf.GetSlice(getUpstreamKey(service, "custom_route")) {
		routes = append(routes, route)
	}
	if len(routes) == 0 { // 兼容老的 不会合并
		for _, route := range qconf.GetSlice("qms.customRoute." + service) {
			routes = append(routes, route)
		}
	}
	for _, route := range routes {
		customRoute := &CustomRoute{}
		routeM := cast.ToStringMap(route)
		if value, ok := routeM["address"]; ok {
			customRoute.Address = cast.ToString(value)
		}
		if value, ok := routeM["weight"]; ok {
			customRoute.Weight = cast.ToInt(value)
		}

		upstream.CustomRoute = append(upstream.CustomRoute, *customRoute)
	}

	// 自定义limit[api相关]
	prefix := Limit + "."
	limit := upstream.RateLimit.Limit
	for key, value := range qconf.GetStringMap(getUpstreamKey(service, "rate_limit")) { // from upstreams
		if strings.HasPrefix(key, prefix) {
			limit[key] = cast.ToInt(value)
		}
	}
	if qconf.GetBool("qms.flowcontrol.Consumer.qps.enabled") {
		prefix = Limit + "." + service + "."
		limit[Limit] = qconf.GetInt(getUpstreamKey(service, "rate_limit.limit"), qconf.GetInt("qms.flowcontrol.Consumer.qps.limit."+service, limit[Limit])) // 可能有优先级问题 需要重新赋值
		for key, value := range qconf.GetStringMap("qms.flowcontrol.Consumer.qps") {
			if strings.HasPrefix(key, prefix) {
				if _, ok := limit[key]; !ok {
					limit[key] = cast.ToInt(value)
				}
			}
		}
	}

	return upstream
}

func defaultUpstream(service string) Upstream {
	if service == constutil.Common {
		return Upstream{
			Sidecar: Sidecar{
				Enabled: defaultSidecarEnabled,
				Address: defaultSidecarAddress,
			},
			Transport: Transport{
				TimeoutMs:     defaultTimeoutMs,
				Loadbalance:   defaultLoadbalance,
				MaxConcurrent: defaultMaxConcurrent,
				MaxIdleConn:   defaultMaxIdleConn,
			},
			Discoverer: Discoverer{
				Tags: make([]string, 0),
			},
			//Retry: Retry{
			//	Condition: defaultRetryCondition,
			//	Backoff: Backoff{
			//		Kind: defaultBackoffKind,
			//	},
			//},
			RateLimit: RateLimit{
				Limit: map[string]int{
					Limit: defaultMaxQPS,
				},
			},
			CircuitBreaker: CircuitBreaker{
				Scope:                    defaultCircuitScope,
				SleepWindowMs:            defaultCircuitSleepWindowMs,
				RequestVolumeThreshold:   defaultCircuitRequestVolumeThreshold,
				ErrorThresholdPercentage: defaultCircuitErrorThresholdPercentage,
			},
			Parampath:   make([]string, 0),
			CustomRoute: make([]CustomRoute, 0),
			Env:         qenv.Get(),
			LoadBalance: LoadBalance{
				Strategy:       defaultLoadBalanceStrategy,
				WithoutBreaker: defaultLoadBalanceWithoutBreaker,
			},
		}
	}

	// service
	upstream := GetUpstream(constutil.Common).copy()
	upstream.Address = service
	return upstream
}

// inheritMessy 继承凌乱的配置
func inheritMessy() {
	added := make(map[string]*Upstream)

	getone := func(sname string) (*Upstream, bool) {
		if isExistUpstream(sname) {
			return nil, false
		}
		upstream, ok := added[sname]
		if !ok {
			tmp := defaultUpstream(sname)
			upstream = &tmp
			added[sname] = upstream
		}
		return upstream, true
	}

	// limiter
	if qconf.GetBool("qms.flowcontrol.Consumer.qps.enabled", false) {
		for key, value := range cast.ToStringMap(searchAdvance("qms.flowcontrol.Consumer.qps")) { // key需要区分大小写 不然APPID会有问题
			keys := strings.SplitN(key, ".", 3) // keys [limit, service, [path]]
			if len(keys) >= 2 {
				upstream, ok := getone(keys[1])
				if !ok {
					continue
				}
				upstream.RateLimit.Enabled = true
				if len(keys) == 2 {
					upstream.RateLimit.Limit["limit"] = cast.ToInt(value)
					continue
				}
				upstream.RateLimit.Limit["limit."+strings.ToLower(keys[2])] = cast.ToInt(value)
			}
		}
	}

	// isolation - transport
	for service, value := range cast.ToStringMap(searchAdvance("qms.isolation.Consumer")) { // key需要区分大小写 不然APPID会有问题
		if transportM, err := cast.ToStringMapIntE(value); err == nil {
			upstream, ok := getone(service)
			if !ok {
				continue
			}
			if timeoutMs, ok := transportM["timeoutInMilliseconds"]; ok {
				upstream.Transport.TimeoutMs = timeoutMs
			}
			if maxConcurrent, ok := transportM["maxConcurrentRequests"]; ok {
				upstream.Transport.MaxConcurrent = maxConcurrent
			}
			if respCacheMs, ok := transportM["respCacheMs"]; ok {
				upstream.Transport.RespCacheMs = respCacheMs
			}
		}
	}

	// circuitBreaker
	for service, value := range cast.ToStringMap(searchAdvance("qms.circuitBreaker.Consumer")) { // key需要区分大小写 不然APPID会有问题
		if circuitM, err := cast.ToStringMapE(value); err == nil {
			upstream, ok := getone(service)
			if !ok {
				continue
			}
			if enabled, ok := circuitM["enabled"]; ok {
				upstream.CircuitBreaker.Enabled = cast.ToBool(enabled)
			}
			if forceOpen, ok := circuitM["forceOpen"]; ok {
				upstream.CircuitBreaker.ForceOpen = cast.ToBool(forceOpen)
			}
			if forceClosed, ok := circuitM["forceClosed"]; ok {
				upstream.CircuitBreaker.ForceClosed = cast.ToBool(forceClosed)
			}
			if sleepMs, ok := circuitM["sleepWindowInMilliseconds"]; ok {
				upstream.CircuitBreaker.SleepWindowMs = cast.ToInt(sleepMs)
			}
			if requestVolumeThreshold, ok := circuitM["requestVolumeThreshold"]; ok {
				upstream.CircuitBreaker.RequestVolumeThreshold = cast.ToInt(requestVolumeThreshold)
			}
			if errorThresholdPercentage, ok := circuitM["errorThresholdPercentage"]; ok {
				upstream.CircuitBreaker.ErrorThresholdPercentage = cast.ToInt(errorThresholdPercentage)
			}
		}
	}

	// metrics-parampath
	for _, path := range cast.ToStringSlice(searchAdvance("qms.metrics.Consumer.parampath")) { // key需要区分大小写 不然APPID会有问题
		list := strings.SplitN(path, "/", 2)
		if len(list) < 2 {
			continue
		}
		upstream, ok := getone(list[0])
		if !ok {
			continue
		}
		upstream.Parampath = append(upstream.Parampath, "/"+list[1])
	}

	// custom_route
	for service, routes := range cast.ToStringMap(searchAdvance("qms.customRoute")) { // key需要区分大小写 不然APPID会有问题
		upstream, ok := getone(service)
		if !ok {
			continue
		}
		for _, route := range cast.ToSlice(routes) {
			customRoute := &CustomRoute{}
			routeM := cast.ToStringMap(route)
			if value, ok := routeM["address"]; ok {
				customRoute.Address = cast.ToString(value)
			}
			if value, ok := routeM["weight"]; ok {
				customRoute.Weight = cast.ToInt(value)
			}
			if customRoute != nil {
				upstream.CustomRoute = append(upstream.CustomRoute, *customRoute)
			}
		}
	}

	// tail
	for key, value := range added {
		conf.Qms.Upstreams[key] = *value
	}
}

func isExistUpstream(name string) (ok bool) {
	_, ok = conf.Qms.Upstreams[name]
	if ok {
		return
	}
	_, ok = conf.Qms.upstreamAlias[name]
	return
}

func getRealName(alias string) string {
	name, ok := Get().upstreamAlias[alias]
	if ok {
		return name
	}
	return alias
}

func getUpstreamKey(service, suffix string) string {
	return "qms.upstreams." + service + "." + suffix
}
