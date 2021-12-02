package grpc

import (
	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/engine"
	servergrpc "git.qutoutiao.net/gopher/qms/internal/server/grpc"
	"google.golang.org/grpc"
)

type ServerOption = base.OptionFunc

type serverOptions struct {
	opts []ServerOption
}

//WithServerName you can specify a unique server name.
func WithServerName(serverName string) ServerOption {
	return engine.WithServerName(serverName)
}

// WithUnaryServerChain 由多个Unary拦截器组成的Chain
func WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return servergrpc.WithUnaryServerChain(interceptors...)
}

// WithStreamServerChain 由多个Stream拦截器组成的Chain
func WithStreamServerChain(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return servergrpc.WithStreamServerChain(interceptors...)
}

// WithServiceDesc 指定grpc服务描述，由pb生成使用
func WithServiceDesc(svcDesc *grpc.ServiceDesc) ServerOption {
	return servergrpc.WithServiceDesc(svcDesc)
}

// WithServer 指定Server, 多grpc共用一个Server场景, 注意要单独使用
func WithServer(s *Server) ServerOption {
	return func(opt interface{}) {
		if o, ok := opt.(*serverOptions); ok {
			if s != nil {
				o.opts = s.opts
			}
		}
	}
}
