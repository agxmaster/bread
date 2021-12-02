package grpc

import (
	"time"

	"git.qutoutiao.net/gopher/qms/internal/base"
	clientgrpc "git.qutoutiao.net/gopher/qms/internal/client/grpc"
	invokergrpc "git.qutoutiao.net/gopher/qms/internal/invoker/grpc"
	"google.golang.org/grpc/metadata"
)

const RouteSidecar = "sidecar" //指定采用sidecar

type CallOption = base.OptionFunc

// Header returns a CallOptions that retrieves the header metadata
// for a unary RPC.
func Header(md *metadata.MD) CallOption {
	return clientgrpc.Header(md)
}

// Trailer returns a CallOptions that retrieves the trailer metadata
// for a unary RPC.
func Trailer(md *metadata.MD) CallOption {
	return clientgrpc.Trailer(md)
}

// WithTimeout 配置请求超时时间
func WithTimeout(timeout time.Duration) CallOption {
	return invokergrpc.WithTimeout(timeout)
}

// WithSidecar用于明确指定采用sidecar代理的方式访问远端
func WithSidecar() CallOption {
	return invokergrpc.WithRoute(RouteSidecar)
}
