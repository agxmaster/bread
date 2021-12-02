package degrade

import (
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/gopher/qudiscovery"
	"github.com/micro/go-micro/registry"
)

// Node 节点数据
type Node struct {
	Address     string
	Name        string
	Host        string
	Port        string
	Weight      int
	IsOffline   bool
	HealthCheck *qudiscovery.HealthCheck
	Node        interface{} // 由熔断器放入 balancer 的数据，可由 nest() 方法取出，数据结构可以自定义。rpc 建议使用 *RpcNextNode ，http 建议使用 *HttpNextNode
}

type HttpNextNode struct {
	ID  string
	URL *url.URL
}

type RpcNextNode registry.Node

func (n *Node) GetSelfProtectionID() string {
	if len(n.Name) > 0 && len(n.Host) > 0 && len(n.Port) > 0 {
		return n.Name + "|" + n.Host + "|" + n.Port
	}
	return n.Address
}

// degrade 状态
type (
	// degrade 状态
	degradeStatus int
	// 数据处理完成后，向下传递  ;   第二个参数用于表示： 是否陷入恐慌状态
	UpdateListData func(DataList []*Node, IsPanic bool)
)

const (
	statusNormal         degradeStatus = 0
	statusSelfProtection degradeStatus = 1 // 自我保护状态, 进入后使用15分钟前全量列表+当前 consul 列表后能通过hc的节点
	statusPanic          degradeStatus = 2 // 恐慌状态, 进入后使用15分钟前全量列表+当前 consul 列表

	degradeFlagOpen  = 0
	degradeFlagClose = 1

	addEndpoints = 1 // 计算降级时先增加的节点数，即默认允许下线的节点数
	//selfProtectionLoopMaxNum = 3 // 自我保护状态循环健康检查的次数
	selfProtectionAutoStop = -2 // 自我保护超时自动退出

	defaultThreshold                 = 0.8
	defaultPanicThreshold            = 0.1
	defaultThresholdContrastInterval = 15 * time.Minute
	defaultEndpointsSaveInterval     = 1 * time.Minute
	defaultHealthCheckInterval       = 30 * time.Second
	defaultPingTimeout               = 500 * time.Millisecond
	defaultSelfProtectionMaxTime     = time.Hour * 24

	MetaKeyStatus     = "status"
	MetaStatusOffline = "offline"
)

// DegradeOpts 服务降级配置
type DegradeOpts struct {
	Flag                      int32         // 0-开启服务降级; 1-关闭服务降级,使用推送过来的数据, 默认 0
	Threshold                 float64       // 自我保护阈值, consul 当前获取的列表数/以前的列表 低于这个值进入自我保护, hc 后节点和当前实时节点一致时推出
	PanicThreshold            float64       // 恐慌阈值, 在自我保护状态下, hc 后节点/以前的列表 低于这个值进入，当此比例大于 Threshold 后推出到自我保护状态
	ThresholdContrastInterval time.Duration // 和多久前对比计算,同时也是历史节点删除的窗口(默认 15m)
	EndpointsSaveInterval     time.Duration // 历史节点归档记录间隔(默认 1m)
	HealthCheckInterval       time.Duration // 两次全局健康检查最少间隔时间(默认 30s)
	PingTimeout               time.Duration // 健康检查超时时间(默认 500ms) [如果注册中心返回了健康检查配置，则此配置无效]
	SelfProtectionMaxTime     time.Duration // 自我保护状态最长持续时间(默认 24h)
}

type kernel struct {
	opts       atomic.Value // 服务降级配置信息
	degradeMap sync.Map
}

func (k *kernel) getDegrade(serviceName string) *degrade {
	if value, ok := k.degradeMap.Load(serviceName); ok {
		return value.(*degrade)
	}

	d := &degrade{
		cluster:                serviceName,
		status:                 statusNormal,
		latestAdapterList:      []*Node{},
		historyEndpoints:       newLoopList(),
		stableHistoryEndpoints: newLoopList(),
		applyList:              []*Node{},
		degradeClose:           make(chan struct{}),
	}
	k.degradeMap.Store(serviceName, d)
	return d
}

// UpdateDegradeOpts 动态更新服务降级配置
func (k *kernel) UpdateDegradeOpts(opts DegradeOpts) {
	if opts.Flag != degradeFlagOpen && opts.Flag != degradeFlagClose {
		opts.Flag = degradeFlagOpen
	}
	if opts.Threshold <= 0 {
		opts.Threshold = defaultThreshold
	}
	if opts.PanicThreshold <= 0 {
		opts.PanicThreshold = defaultPanicThreshold
	}
	if opts.ThresholdContrastInterval <= 0 {
		opts.ThresholdContrastInterval = defaultThresholdContrastInterval
	}
	if opts.EndpointsSaveInterval <= 0 {
		opts.EndpointsSaveInterval = defaultEndpointsSaveInterval
	}
	if opts.HealthCheckInterval <= 0 {
		opts.HealthCheckInterval = defaultHealthCheckInterval
	}
	if opts.PingTimeout <= 0 {
		opts.PingTimeout = defaultPingTimeout
	}
	if opts.SelfProtectionMaxTime <= 0 {
		opts.SelfProtectionMaxTime = defaultSelfProtectionMaxTime
	}

	k.opts.Store(opts)

	// 当关闭降级开关后，立即停止所有检查，并回退到正常模式
	if opts.Flag == degradeFlagClose {
		k.degradeMap.Range(func(key, value interface{}) bool {
			// 恢复 UpstreamDiffer
			//value.(*degrade).recoverUpstreamDiffer()

			select {
			case value.(*degrade).degradeClose <- struct{}{}:
				//fmt.Println(key, "降级关闭") // 仅用于测试
			default:
			}
			return true
		})
	}
}

// GetDegradeOpts 获取服务降级配置
func (k *kernel) GetDegradeOpts() DegradeOpts {
	return k.opts.Load().(DegradeOpts)
}
