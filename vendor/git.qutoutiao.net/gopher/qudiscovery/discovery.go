package qudiscovery

import (
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	// ErrServiceNameInvalid 服务名格式错误
	ErrServiceNameInvalid = errors.New("Service name invalid")

	// HealthcheckMetaKey 健康检查信息 meta 字段的 key
	HealthcheckMetaKey = "healthcheck"
	RegtimeMetaKey     = "regtime"

	regexpService     = regexp.MustCompile("^([0-9,a-z,A-Z,-]+\\.)*([0-9,a-z,A-Z,-]+)\\.service\\.discovery$")
	regexpServiceName = regexp.MustCompile("^([0-9,a-z,A-Z,-]+)$")
)

// OnUpdate 请不要修改任何 *Event 中的变量
type OnUpdate func(e *Event)

// OnUpdateList 请不要修改任何 *Service 中的变量
type OnUpdateList func(*ServiceList)

// Adapter 注册中心具体实现
type Adapter interface {
	// Register 注册一个服务
	Register(svc *Service, opts ...RegisterOpts) (Registry, error)
	// WatchList 监听一个服务变更一次性返回整个 list, 不区分 tag, 直接监听所有 tags
	WatchList(service string, onUpdateList OnUpdateList, opts ...WatchOption) (Watcher, error)
}

// Discovery 服务发现
type Discovery interface {
	Adapter
	// Watch 监听一个服务变更, 不区分 tag, 直接监听所有 tags
	Watch(service string, onUpdate OnUpdate, opts ...WatchOption) (Watcher, error)
	// GetServers 按照这组服务节点列表和版本号, 请不要修改任何 *ServiceList 中的变量,tag被废弃不被用作prd/pre环境区分(现在通过服务名来区分环境)
	GetServers(service string, tags ...string) (*ServiceList, error)
	GetServersWithDC(service string, dc string, tags ...string) (*ServiceList, error)
	// UpdateDegradeOpts 动态更新降级配置
	UpdateDegradeOpts(opts DegradeOpts) error
	// Shutdown 停止
	Shutdown() error
}

// Registry 服务注册后用于注销或更新
type Registry interface {
	// Deregister 注销服务
	Deregister() error
	// Update 更新服务注册信息
	Update(svc *Service, opts ...RegisterOpts) error
}

// Watcher 单个服务事件监听
type Watcher interface {
	Stop() error
	WaitReady() error
}

// ServiceList 服务节点列表
type ServiceList struct {
	WatchVersion uint64  // watcher 版本号(递增)
	Version      uint64  // 服务列表版本号(递增),[0,maxUint32]表示不走降级,[maxUint32+1,maxUint64]表示降级列表,版本号有可能从高到底转换
	RunMode      RunMode // 运行模式，正常或者恢复到了历史版本,变更会刷新缓存(只在 wrap 被使用)
	Services     []*Service
}

// Action 变更事件
type Action int

const (
	// ActionAdd 增加节点
	ActionAdd Action = 1
	// ActionDel 减少节点
	ActionDel Action = 2
	// ActionMod 修改节点
	ActionMod Action = 3
)

// RunMode 发现库运行模式
type RunMode int32

const (
	// RunModeNormal 正常的服务发现
	RunModeNormal RunMode = 1
	// RunModeRecover 使用本地文件
	RunModeRecover RunMode = 2
	// RunModeSelfProtection 节点处于自我保护模式
	RunModeSelfProtection RunMode = 3
	// RunModeInit 节点处于初始化数据
	RunModeInit RunMode = 4
	// RunModePanic 节点处于恐慌状态
	RunModePanic RunMode = 5
)

// Event 服务事件
type Event struct {
	Action  Action
	Service *Service
}

// Service 服务信息
type Service struct {
	ID          string            `json:"id"`           // 节点全局唯一标识
	Name        string            `json:"name"`         // 服务名
	Address     string            `json:"address"`      // 节点地址
	Port        int               `json:"port"`         // 节点端口
	Tags        []string          `json:"tags"`         // 节点 tag 属性
	Weight      int               `json:"weight"`       // 节点权重
	Meta        map[string]string `json:"meta"`         // metadata
	HealthCheck *HealthCheck      `json:"health_check"` // 注册到 consul 中的健康检查(返回节点时才会存在,注册时填写无效)
}

func (s *Service) GetSelfProtectionID() string {
	if len(s.Name) > 0 && len(s.Address) > 0 && s.Port > 0 {
		return s.Name + "|" + s.Address + "|" + strconv.FormatInt(int64(s.Port), 10)
	}
	Log.Errorf("Service data error [%+v]", *s)
	return s.ID
}

// RegisterOpts 服务注册配置
type RegisterOpts struct {
	CheckTTL      time.Duration // 服务注册后和 agent 心跳间隔(注册中心异常后可检测到), 默认为 Discovery 中的配置
	CheckIP       string        // 健康检查 IP, 默认为注册的服务 IP
	CheckPort     int           // 健康检查 tcp 端口, 默认为注册的服务端口
	CheckInterval time.Duration // 健康检查间隔时间, 默认 5s
	CheckTimeout  time.Duration // 健康检查超时时间, 默认 2s
	CheckHTTP     *HTTPCheck    // http 健康检查,为空则使用 tcp 端口检查
}

// WatchOption watch配置
type WatchOption struct {
	DC string
}

// HealthCheck 返回的健康检查
type HealthCheck struct {
	Interval api.ReadableDuration `json:"interval,omitempty"` // 健康检查间隔时间
	Timeout  api.ReadableDuration `json:"timeout,omitempty"`  // 健康检查超时时间
	HTTP     string               `json:"http,omitempty"`     // http 健康检查完整路径 http://...
	Header   map[string][]string  `json:"header,omitempty"`   // http 健康检查带上的头
	Method   string               `json:"method,omitempty"`   // http 健康检查的Method
	TCP      string               `json:"tcp,omitempty"`      // tcp 健康检查的 ip:port
}

// HTTPCheck http 健康检查
type HTTPCheck struct {
	Path   string // 路径
	Method string
	Header map[string][]string
}

// DegradeOpts 降级配置
type DegradeOpts struct {
	Threshold                 float64       // 自我保护阈值, consul 当前获取的列表数/以前的列表 低于这个值进入自我保护, hc 后节点和当前实时节点一致时推出
	PanicThreshold            float64       // 恐慌阈值, 在自我保护状态下, hc 后节点/以前的列表 低于这个值进入，当此比例大于 Threshold 后推出到自我保护状态
	ThresholdContrastInterval time.Duration // 和多久前对比计算,同时也是历史节点删除的窗口(默认 15m)
	EndpointsSaveInterval     time.Duration // 历史节点归档记录间隔(默认 1m)
	Flag                      int32         // 0-开启自我保护; 1-关闭自我保护,使用注册中心的数据, 默认 0
	HealthCheckInterval       time.Duration // 两次全局健康检查最少间隔时间,默认 30s
	PingTimeout               time.Duration // 健康检查超时时间,默认 500ms(如果注册中心返回了健康检查配置，则此配置无效)
}

// Logger 日志
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Writer() *io.PipeWriter
}

// GetTagAndServiceName 通过域名获取服务名和 tag
// Deprecated: 已经弃用,请使用 GetTagsAndServiceName()
func GetTagAndServiceName(serviceDomain string) (service, tag string, err error) {
	tmps := regexpService.FindSubmatch([]byte(serviceDomain))
	switch len(tmps) {
	case 3:
		if len(tmps[1]) > 0 {
			tmps[1] = tmps[1][:len(tmps[1])-1]
		}
		return string(tmps[2]), string(tmps[1]), nil
	default:
		err = ErrServiceNameInvalid
		return
	}
}

// GetTagsAndServiceName 通过域名获取服务名和 tags
func GetTagsAndServiceName(serviceDomain string) (service string, tags []string, err error) {
	if !regexpService.MatchString(serviceDomain) {
		err = ErrServiceNameInvalid
		return
	}
	tmps := strings.Split(serviceDomain, ".")
	if len(tmps) < 3 {
		err = ErrServiceNameInvalid
		return
	}

	tags = tmps[:len(tmps)-3]
	service = tmps[len(tmps)-3]
	return
}

// CheckServiceOrTagName 检查服务名或者tag是否符合规范
func CheckServiceOrTagName(name string) bool {
	return regexpServiceName.MatchString(name)
}

// CopyService 复制 Service 对象
func CopyService(svc *Service) *Service {
	ret := *svc
	if svc.Tags != nil {
		tags := make([]string, len(svc.Tags))
		copy(tags, svc.Tags)
		ret.Tags = tags
	}
	if svc.Meta != nil {
		meta := make(map[string]string)
		for k, v := range svc.Meta {
			meta[k] = v
		}
		ret.Meta = meta
	}
	if svc.HealthCheck != nil {
		newHealthCheck := *svc.HealthCheck
		if len(svc.HealthCheck.Header) > 0 {
			header := make(map[string][]string)
			for k, v := range svc.HealthCheck.Header {
				newl := make([]string, len(v))
				copy(newl, v)
				header[k] = v
			}
			newHealthCheck.Header = header
		}
		ret.HealthCheck = &newHealthCheck
	}
	return &ret
}
