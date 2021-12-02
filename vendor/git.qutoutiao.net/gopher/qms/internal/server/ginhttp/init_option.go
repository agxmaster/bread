package ginhttp

import (
	"crypto/tls"
	"net"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/server"
)

type HcFunc func() map[string]string

type initOption struct {
	serverName string
	address    string
	listen     net.Listener
	tLSConfig  *tls.Config
	chainName  string
	bodyLimit  int64
	extraHc    HcFunc
}

func newInitOption(opts *server.InitOptions) *initOption {
	opt := &initOption{
		serverName: opts.ServerName,
		address:    opts.Address,
		listen:     opts.Listen,
		tLSConfig:  opts.TLSConfig,
		chainName:  common.DefaultChainName,
		bodyLimit:  config.GlobalDefinition.Qms.Transport.MaxBodyBytesMap["rest"],
	}
	for _, o := range opts.Opts {
		o(opt)
	}

	return opt
}

// WithConfigDir is option to set extra hc func.
// NOTE: 如果配置文件中开启了hc，框架内部已经实现了hc接口和基本的检测，使用方可以通过此方法额外添加检测项。
func WithExtraHc(hc HcFunc) base.OptionFunc {
	return func(opt interface{}) {
		if o, ok := opt.(*initOption); ok {
			if hc != nil {
				o.extraHc = hc
			}
		}
	}
}
