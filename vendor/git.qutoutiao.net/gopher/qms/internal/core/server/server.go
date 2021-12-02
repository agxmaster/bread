//Package server is a package for protocol of a micro service
package server

import "git.qutoutiao.net/gopher/qms/internal/base"

type Option = base.OptionFunc

// Server interface for the protocol server, a server should implement init, register, start, and stop
type Server interface {
	Register(interface{}, ...Option) (string, error)
	Start() error
	Stop() error
	String() string
}

// GinServer interface for the gin server
type GinServer interface {
	Server
	//Engine return the *gin.Engine
	Engine() interface{}
}
