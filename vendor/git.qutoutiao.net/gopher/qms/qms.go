package qms

import (
	// engine first
	"git.qutoutiao.net/gopher/qms/internal/engine"

	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"git.qutoutiao.net/gopher/qms/protocol/grpc"
	"git.qutoutiao.net/gopher/qms/protocol/rest"
	"github.com/gin-gonic/gin"
)

//Init 初始化框架内部的配置与各个组件，是框架相关功能的前置条件。
func Init(options ...InitOption) error {
	return engine.Init(options...)
}

//Run 启动服务、监听端口、并阻塞于信号监听，收到退出信号后，会优雅退出。
func Run(options ...RunOption) error {
	return engine.Run(options...)
}

//GraceFork 通过封装程序的main入口函数，实现master/slave方式的平滑重启。
func GraceFork(main func()) {
	engine.GraceFork(main)
}

//Gin return a *gin.Engine that you can register route with gin
//
//Deprecated: use rest.Gin instead.
func Gin(opts ...rest.ServerOption) (*gin.Engine, error) {
	return rest.Gin(opts...)
}

//RegisterSchema Register a API service to specific server by name.
//
//Deprecated: use grpc.RegisterSchema instead.
func RegisterSchema(defaultServerName string, structPtr interface{}, opts ...server.Option) {
	grpc.RegisterSchema(defaultServerName, structPtr, opts...)
}
