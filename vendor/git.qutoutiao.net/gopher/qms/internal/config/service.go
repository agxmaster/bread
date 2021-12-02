package config

import (
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"github.com/spf13/cast"
)

// service相关

type Service struct {
	AppID    string
	Env      string
	Version  string
	ServiceM map[string]ServiceSpec
}

type ServiceSpec struct {
	Address        string
	Protocol       protocol.Protocol
	GrpcurlEnabled bool
	Registrator    Registrator
	RateLimit      RateLimit
}

type Registrator struct {
	Disabled bool
	Tags     []string
}

func (s ServiceSpec) copy() ServiceSpec {
	sspec := s

	sspec.Registrator.Tags = make([]string, len(s.Registrator.Tags))
	copy(sspec.Registrator.Tags, s.Registrator.Tags)

	limit := make(map[string]int)
	for key, value := range s.RateLimit.Limit {
		limit[key] = value
	}
	sspec.RateLimit.Limit = limit

	return sspec
}

// TODO: 这个地方可能要加上app_id前缀了
func GetService(name string) *ServiceSpec {
	serviceMap := Get().Service.ServiceM

	value, ok := serviceMap[name]
	if !ok {
		value = serviceMap[constutil.Common]
	}
	return &value
}

// GetServiceMap 不包含common
func GetServiceMap() map[string]ServiceSpec {
	serviceMap := make(map[string]ServiceSpec, len(Get().Service.ServiceM))
	for name, spec := range Get().Service.ServiceM {
		if name != constutil.Common {
			serviceMap[name] = spec.copy()
		}
	}
	return serviceMap
}

func initService() {
	// TODO: 关于参数的校验 需要考虑下[不同模式 应该有不同的校验 required也是动态的]
	conf.Qms.Service = Service{
		AppID:   qconf.GetString("qms.service.app_id", qconf.GetString("service.name")),                 // SDK模式可以没有Name
		Env:     qconf.GetString("qms.service.env", qconf.GetString("service.environment", defaultEnv)), // SDK模式可以没有Env
		Version: qconf.GetString("qms.service.version", qconf.GetString("service.version", defaultVersion)),
		ServiceM: map[string]ServiceSpec{
			constutil.Common: newServiceSpec(constutil.Common, qconf.GetQconf()),
		},
	}

	for service, value := range qconf.GetStringMap("qms.service", qconf.GetStringMap("qms.protocols")) {
		if service == constutil.Common {
			continue
		}
		if _, err := cast.ToStringMapE(value); err == nil {
			conf.Qms.Service.ServiceM[service] = newServiceSpec(service, qconf.GetQconf())
		}
	}

	return
}

func newServiceSpec(service string, qconf *qconf.Qconf) ServiceSpec {
	sspec := defaultServiceSpec(service)
	sspec.Registrator = Registrator{
		Disabled: qconf.GetBool(getServiceKey(service, "registrator.disabled"), sspec.Registrator.Disabled),
		Tags:     qconf.GetStringSlice(getServiceKey(service, "registrator.tags"), sspec.Registrator.Tags),
	}
	sspec.RateLimit = RateLimit{
		Enabled: qconf.GetBool(getServiceKey(service, "rate_limit.enabled"), sspec.RateLimit.Enabled),
		Limit: map[string]int{
			Limit: qconf.GetInt(getServiceKey(service, "rate_limit.limit"), sspec.RateLimit.Limit[Limit]),
		},
	}
	if service == constutil.Common {
		sspec.Registrator.Disabled = qconf.GetBool(getServiceKey(service, "registrator.disabled"), qconf.GetBool("qms.service.registry.registerDisabled", sspec.Registrator.Disabled))
		sspec.RateLimit.Enabled = qconf.GetBool(getServiceKey(service, "rate_limit.enabled"), qconf.GetBool("qms.flowcontrol.Provider.qps.enabled", sspec.RateLimit.Enabled))
		sspec.RateLimit.Limit[Limit] = qconf.GetInt(getServiceKey(service, "rate_limit.limit"), qconf.GetInt("qms.flowcontrol.Provider.qps.global.limit", sspec.RateLimit.Limit[Limit]))
		return sspec
	}

	// service独有的[address protocol等]
	sspec.Address = qconf.GetString(getServiceKey(service, "address"), qconf.GetString(getkey("qms.protocols", service, "listenAddress"), sspec.Address))
	sspec.Protocol = getProtocolByName(service)
	sspec.GrpcurlEnabled = qconf.GetBool(getServiceKey(service, "grpcurl_enabled"), qconf.GetBool(getkey("qms.protocols", service, "enableGrpcurl"), sspec.GrpcurlEnabled))

	// 自定义limit[api相关]
	prefix := Limit + "."
	limit := sspec.RateLimit.Limit
	for key, value := range qconf.GetStringMap(getServiceKey(service, "rate_limit")) { // from upstreams
		if strings.HasPrefix(key, prefix) {
			limit[key] = cast.ToInt(value)
		}
	}
	if qconf.GetBool("qms.flowcontrol.Provider.qps.enabled") {
		prefix = Limit + "."
		for key, value := range qconf.GetStringMap("qms.flowcontrol.Provider.qps") {
			if strings.HasPrefix(key, prefix) {
				if _, ok := limit[key]; !ok { // 不存在 才添加
					limit[key] = cast.ToInt(value)
				}
			}
		}
	}

	return sspec
}

// 获取默认ServiceSpec配置
func defaultServiceSpec(service string) ServiceSpec {
	if service == constutil.Common { // 给予默认值
		return ServiceSpec{
			Registrator: Registrator{
				Tags: make([]string, 0),
			},
			RateLimit: RateLimit{
				Limit: map[string]int{
					Limit: defaultMaxQPS,
				},
			},
		}
	}

	// service
	return Get().Service.ServiceM[constutil.Common].copy()
}
