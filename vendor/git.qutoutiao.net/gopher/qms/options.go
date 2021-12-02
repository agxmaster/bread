package qms

import (
	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/engine"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/internal/server/ginhttp"
)

type InitOption = engine.InitOption
type CustomMetricLabel = metrics.CustomLabel

// WithConfigDir is option to set config dir.
// NOTE: 有些场景可能需要明确指定配置文件目录，例如:单元测试时。
func WithConfigDir(dir string) InitOption {
	return engine.WithConfigDir(dir)
}

// WithConfigDir is option to set extra hc func.
// NOTE: 如果配置文件中开启了hc，框架内部已经实现了hc接口和基本的检测，使用方可以通过此方法额外添加检测项。
func WithExtraHc(hc func() map[string]string) InitOption {
	return ginhttp.WithExtraHc(ginhttp.HcFunc(hc))
}

// WithMetricsLabel 用于设置自定义的QPS metric label
// NOTE: 目前只针对http服务有效
func WithMetricLabels(labels ...CustomMetricLabel) InitOption {
	return engine.WithMetricLabels(labels...)
}

type RunOption = engine.RunOption

// WithExitCallback is option to set exit callback for graceful shutdown.
// NOTE: 优雅退出时，如果业务上也有逻辑需要优雅退出，则可以在此回掉函数中来实现。
func WithExitCallback(cb func()) RunOption {
	return engine.WithExitCallback(cb)
}

//Deprecated: use grpc/rest.ServerOption in qms/protocol/grpc or qms/protocol/rest package instead.
type ServerOption = engine.ServerOption

//Deprecated: use grpc/rest.DialOption in qms/protocol/grpc or qms/protocol/rest package instead.
type DialOption = client.DialOption

//Deprecated: use grpc/rest.CallOption in qms/protocol/grpc or qms/protocol/rest package instead.
type CallOption = client.CallOption

//WithServerName you can specify a unique server name.
//
//Deprecated: use grpc/rest.WithServerName in qms/protocol/grpc or qms/protocol/rest package instead.
func WithServerName(serverName string) ServerOption {
	return engine.WithServerName(serverName)
}
