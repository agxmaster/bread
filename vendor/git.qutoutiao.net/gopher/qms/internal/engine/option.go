package engine

import (
	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

type InitOption = base.OptionFunc

// initOptions struct having information about init parameters
type initOptions struct {
	configDir    string
	metricLabels []metrics.CustomLabel
}

// WithConfigDir is option to set config dir.
func WithConfigDir(dir string) InitOption {
	return func(options interface{}) {
		if o, ok := options.(*initOptions); ok {
			o.configDir = dir
		}
	}
}

// WithMetricLabels is option to set custom metric labels.
func WithMetricLabels(labels ...metrics.CustomLabel) InitOption {
	return func(options interface{}) {
		if o, ok := options.(*initOptions); ok {
			o.metricLabels = labels
		}
	}
}

// WithEnv 主要是提供给SDK模式设置环境变量，从而初始化默认配置
func WithEnv(env qenv.Env) InitOption {
	return func(options interface{}) {
		if env.IsValid() {
			qconf.Set("qms.service.env", env.String())
		}
	}
}

// RunOptions struct having information about run parameters
type runOptions struct {
	exitCb ExitCallback
}

type RunOption base.OptionFunc

//ExitCallback is a function which would be invoked when service shutdown
type ExitCallback func()

// WithExitCallback is option to set exit callback for graceful shutdown.
func WithExitCallback(cb ExitCallback) RunOption {
	return func(opts interface{}) {
		if o, ok := opts.(*runOptions); ok {
			o.exitCb = cb
		}
	}
}

type serverOptions struct {
	serverName string
}

type ServerOption = base.OptionFunc

func WithServerName(serverName string) ServerOption {
	return func(opts interface{}) {
		if o, ok := opts.(*serverOptions); ok {
			if serverName != "" {
				o.serverName = serverName
			}
		}
	}
}
