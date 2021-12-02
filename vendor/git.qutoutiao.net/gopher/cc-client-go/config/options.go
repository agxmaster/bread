package config

import (
	"time"

	"git.qutoutiao.net/gopher/cc-client-go"
	"git.qutoutiao.net/gopher/cc-client-go/log"
)

type Option func(c *Options)

type Options struct {
	// 业务日志，目前是key没有找到的日志，默认包含上报logger
	// 索引0为自定义logger，索引1为上报到配置中心的logger
	appLoggers [2]log.Logger
	// 诊断日志, 默认包含上报logger
	diagnosticLogger [2]log.Logger
	// 加载备份文件，当初始化拉取失败
	restoreWhenInitFail bool
	// 回调
	onChange func(*cc.ConfigCenter) error
	// 备份路径
	backupDir string

	// 最大重试间隔
	maxInterval time.Duration
	// 初始化重试间隔
	initialInterval time.Duration
	// 重试随机因子
	jitterFraction float64

	QAServerURL string
	// 开启debug
	Debug bool
	// 开发环境配置文件读取
	DevConfigFilePath string
}

func AppLogger(logger log.Logger) Option {
	return func(c *Options) {
		c.appLoggers[0] = logger
	}
}

func DiagnosticLogger(logger log.Logger) Option {
	return func(c *Options) {
		c.diagnosticLogger[0] = logger
	}
}

func RestoreWhenInitFail(restore bool) Option {
	return func(c *Options) {
		c.restoreWhenInitFail = restore
	}
}

func OnChange(onChange func(*cc.ConfigCenter) error) Option {
	return func(c *Options) {
		c.onChange = onChange
	}
}

func BackupDir(backupDir string) Option {
	return func(c *Options) {
		c.backupDir = backupDir
	}
}
func MaxInterval(maxInterval time.Duration) Option {
	return func(c *Options) {
		c.maxInterval = maxInterval
	}
}

func InitialInterval(initInterval time.Duration) Option {
	return func(c *Options) {
		c.initialInterval = initInterval
	}
}

func JitterFraction(jitterFraction float64) Option {
	return func(c *Options) {
		c.jitterFraction = jitterFraction
	}
}

func QAServerURL(url string) Option {
	return func(c *Options) {
		c.QAServerURL = url
	}
}

func Debug(debug bool) Option {
	return func(c *Options) {
		c.Debug = debug
	}
}

func DevConfigFilePath(path string) Option {
	return func(c *Options) {
		c.DevConfigFilePath = path
	}
}

func applyOptions(options ...Option) Options {
	opts := Options{
		appLoggers:          [2]log.Logger{log.NullLogger, log.NullLogger},
		diagnosticLogger:    [2]log.Logger{log.NullLogger, log.NullLogger},
		restoreWhenInitFail: true,
	}
	for _, option := range options {
		option(&opts)
	}
	return opts
}
