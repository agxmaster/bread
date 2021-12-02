package server

import (
	"crypto/tls"
	"net"

	"git.qutoutiao.net/gopher/qms/internal/base"
)

type InitOptions struct {
	listenM       map[string]net.Listener
	ServerName    string
	Address       string
	Listen        net.Listener
	TLSConfig     *tls.Config
	EnableGrpcurl bool
	Opts          []base.OptionFunc // 外部传递
}

func WithListener(addr string, l net.Listener) base.OptionFunc {
	return func(options interface{}) {
		if o, ok := options.(*InitOptions); ok {
			if o.listenM == nil {
				o.listenM = make(map[string]net.Listener)
			}
			if l != nil {
				o.listenM[addr] = l
			}
		}
	}
}
