package grpc

import (
	"time"

	"git.qutoutiao.net/gopher/qms/balancer"
	"git.qutoutiao.net/gopher/qms/internal/base"
	clientgrpc "git.qutoutiao.net/gopher/qms/internal/client/grpc"
	invokergrpc "git.qutoutiao.net/gopher/qms/internal/invoker/grpc"
	"google.golang.org/grpc"
)

type DialOption = base.OptionFunc

// WithRemoteService 指定要访问的远端服务[consul注册服务名、ip:port、域名]
func WithRemoteService(remoteService string) DialOption {
	return invokergrpc.WithRemoteService(remoteService)
}

// WithBalancerName 使用已注册的LB
func WithBalancerName(name string) DialOption {
	return invokergrpc.WithBalancerName(name)
}

// WithBalancer 注册自定义LB
func WithBalancer(pb balancer.PickerBuilder) DialOption {
	return invokergrpc.WithBalancer(pb)
}

// WithBlock 阻塞等待连接建立
func WithBlock() DialOption {
	return clientgrpc.WithBlock()
}

// WithDialTimeout 指定创建连接的超时时间
func WithDialTimeout(d time.Duration) DialOption {
	return clientgrpc.WithDialTimeout(d)
}

// WithUnaryClientChain 由多个Unary拦截器组成的客户端Chain
func WithUnaryClientChain(interceptors ...grpc.UnaryClientInterceptor) DialOption {
	return clientgrpc.WithUnaryClientChain(interceptors...)
}

// WithStreamClientChain 由多个Stream拦截器组成的客户端Chain
func WithStreamClientChain(interceptors ...grpc.StreamClientInterceptor) DialOption {
	return clientgrpc.WithStreamClientChain(interceptors...)
}
