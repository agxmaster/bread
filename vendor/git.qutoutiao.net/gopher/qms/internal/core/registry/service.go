package registry

import (
	"strconv"

	"git.qutoutiao.net/gopher/qms/internal/core/metadata"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/registryutil"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// Service 服务需要包含的信息
// TODO: 后续可以和server打通
type Service struct {
	ID       string                    // DiscoveryID[注册时不需要填入]
	Name     string                    // 服务名称
	AppID    string                    // PaasID[仅注册]
	IP       string                    // IP地址
	Port     int                       // 端口
	Endpoint string                    // IP:Port
	Weight   int32                     // 权重
	Env      qenv.Env                  // 环境[仅注册]
	Protocol protocol.Protocol         // 协议[仅注册]
	Tags     []string                  // 标签
	Meta     map[string]string         // 元信息
	HC       *registryutil.HealthCheck // HTTP健康检查

	registryName    string
	stdRegistryName string
}

// NewService 创建携带tags、metadata、hc的service
func NewService(port int, proto protocol.Protocol) *Service {
	s := &Service{
		AppID:    runtime.ServiceName,
		IP:       iputil.GetLocalIP(),
		Port:     port,
		Endpoint: iputil.GetLocalIP() + ":" + strconv.Itoa(port),
		Weight:   100,
		Env:      qenv.Get(),
		Protocol: proto,
	}
	s.registryName = toRegistryName(s.AppID, s.Protocol)
	s.stdRegistryName = toStdRegistryName(s.AppID, s.Protocol, s.Env)

	// tags
	framework := metadata.NewFramework()
	s.Tags = []string{
		s.Protocol.String(),
		framework.Name + ":" + framework.Version,
		s.Env.String(),
	}

	// metadata
	s.Meta = map[string]string{
		"protoc":               proto.String(),
		registryutil.AppID:     s.AppID,
		registryutil.Region:    "",
		registryutil.Zone:      "",
		registryutil.Container: registryutil.GetContainer(),
		registryutil.StatusStr: registryutil.Online.String(),
		registryutil.Weight:    "100",
	}

	if proto == protocol.ProtocHTTP {
		// 只有HTTP才创建HealthCheck
		hc := registryutil.NewHealthCheck(s.AppID, s.Endpoint)
		s.Meta[registryutil.Healthcheck] = hc.String()
		s.HC = hc
	}
	return s
}
