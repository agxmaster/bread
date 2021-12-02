package grpc

import (
	"git.qutoutiao.net/gopher/qms/internal/engine"
)

type Server struct {
	opts []ServerOption
}

// RegisterSchema 向engine注册Schema，业务方不需要关心 pb自动生成
func RegisterSchema(defaultServerName string, structPtr interface{}, opts ...ServerOption) {
	var opt serverOptions
	for _, o := range opts {
		o(&opt)
	}
	opts = append(opts, opt.opts...)
	engine.RegisterSchema(defaultServerName, structPtr, opts...)
}

// NewServer 提供ServerOption创建server，多grpc公用一个Server时需要先使用其创建公共Server
func NewServer(opts ...ServerOption) *Server {
	return &Server{
		opts: opts,
	}
}
