package grpc

import (
	"crypto/tls"
	"net"

	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"google.golang.org/grpc"
)

type serverOption struct {
	listen        net.Listener
	address       string
	serverName    string
	chainName     string
	tLSConfig     *tls.Config
	enableGrpcurl bool
	grpcOpts      []grpc.ServerOption
	unaryInts     []grpc.UnaryServerInterceptor
	streamInts    []grpc.StreamServerInterceptor
	svcDesc       *grpc.ServiceDesc
}

func newServerOption(opts *server.InitOptions) *serverOption {
	return &serverOption{
		serverName:    opts.ServerName,
		address:       opts.Address,
		listen:        opts.Listen,
		chainName:     common.DefaultChainName,
		tLSConfig:     opts.TLSConfig,
		enableGrpcurl: opts.EnableGrpcurl,
		unaryInts:     []grpc.UnaryServerInterceptor{wrapUnaryInterceptor(common.DefaultChainName, opts.ServerName)},
		streamInts:    []grpc.StreamServerInterceptor{wrapStreamInterceptor(common.DefaultChainName, opts.ServerName)},
	}
}

func WithUnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) server.Option {
	return func(opt interface{}) {
		if o, ok := opt.(*serverOption); ok {
			o.unaryInts = append(o.unaryInts, interceptors...)
		}
	}
}

func WithStreamServerChain(interceptors ...grpc.StreamServerInterceptor) server.Option {
	return func(opt interface{}) {
		if o, ok := opt.(*serverOption); ok {
			o.streamInts = append(o.streamInts, interceptors...)
		}
	}
}

func WithServiceDesc(svcDesc *grpc.ServiceDesc) server.Option {
	return func(opt interface{}) {
		if o, ok := opt.(*serverOption); ok {
			if svcDesc != nil {
				o.svcDesc = svcDesc
			}
		}
	}
}
