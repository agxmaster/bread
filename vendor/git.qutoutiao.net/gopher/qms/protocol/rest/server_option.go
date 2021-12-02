package rest

import (
	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/engine"
)

type ServerOption = base.OptionFunc

//WithServerName you can specify a unique server name.
func WithServerName(serverName string) ServerOption {
	return engine.WithServerName(serverName)
}
