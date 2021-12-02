package balancer

import (
	"context"

	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/registry"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

type Balancer interface {
	// 更新picker
	OnChanged(key *registry.ServiceKey, instances []*Service) error
	// ReBalancer
	ReBalancer(ctx context.Context, opts *Options) (ReBalancer, error)
}

type Options struct {
	SourceService string            // 本服务
	RemoteService string            // 远程调用服务
	Protocol      protocol.Protocol // rest or grpc
	Env           qenv.Env
	Path          string      // [rest: url-path] [grpc: FullMethod]
	Args          interface{} // [rest: http.Request] [grpc: input]
	Datacenter    string      // 指定DC
	Tags          []string    // 指定Tag
}

func NewOptions(inv *invocation.Invocation) *Options {
	return &Options{
		SourceService: inv.SourceServiceID,
		RemoteService: inv.GetRemoteService(),
		Protocol:      inv.Protocol,
		Env:           inv.Env,
		Path:          inv.OperationID,
		Args:          inv.Args,
		Datacenter:    inv.GetUpstream().Discoverer.Datacenter,
		Tags:          inv.GetUpstream().Discoverer.Tags,
	}
}

// LBError load balance error
type LBError struct {
	Message string
}

// Error for to return load balance error message
func (e LBError) Error() string {
	return "lb: " + e.Message
}
