package rest

import (
	"crypto/tls"

	"git.qutoutiao.net/gopher/qms/internal/base"
)

type dialOption struct {
	poolSize  int
	tLSConfig *tls.Config
}

func newDialOption() *dialOption {
	return &dialOption{
		poolSize: DefaultMaxConnsPerHost,
	}
}

func WithPoolSize(poolSize int) base.OptionFunc {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			if poolSize > 0 {
				o.poolSize = poolSize
			}
		}
	}
}

func WithCredentials(tLSConfig *tls.Config) base.OptionFunc {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			if tLSConfig != nil {
				o.tLSConfig = tLSConfig
			}
		}
	}
}
