package cc

import (
	"time"

	"google.golang.org/grpc/codes"

	"git.qutoutiao.net/gopher/cc-client-go/log"
)

type Options struct {
	diagnosticLogger [2]log.Logger
	// 加载备份文件当初始化拉取失败
	restoreWhenInitFail bool
	// 重试错误码
	retryableCodes []codes.Code
	// backup dir
	backupDir string
	// 回调
	onChange func(*ConfigCenter) error
	// 最大重试间隔
	maxInterval time.Duration
	// 初始化重试间隔
	initialInterval time.Duration
	// 重试随机因子
	jitterFraction float64
	// QA server url
	QAServerURL string
	// 是否开启debug
	debug bool
	// client IP
	clientIP string
	// dev环境本地文件
	DevConfigFilePath string
}

type Option func(*Options)

func DevConfigFilePath(path string) Option {
	return func(options *Options) {
		options.DevConfigFilePath = path
	}
}

func QAServerURL(url string) Option {
	return func(options *Options) {
		options.QAServerURL = url
	}
}

func DiagnosticLogger(loggers [2]log.Logger) Option {
	return func(options *Options) {
		options.diagnosticLogger = loggers
	}
}

func MaxInterval(maxInterval time.Duration) Option {
	return func(options *Options) {
		options.maxInterval = maxInterval
	}
}

func InitialInterval(initialInterval time.Duration) Option {
	return func(options *Options) {
		options.initialInterval = initialInterval
	}
}

func JitterFraction(jitterFraction float64) Option {
	return func(options *Options) {
		options.jitterFraction = jitterFraction
	}
}

func RestoreWhenInitFail(restore bool) Option {
	return func(options *Options) {
		options.restoreWhenInitFail = restore
	}
}

func RetryableCodes(codes []codes.Code) Option {
	return func(options *Options) {
		options.retryableCodes = codes
	}
}

func OnChange(onChange func(*ConfigCenter) error) Option {
	return func(options *Options) {
		options.onChange = onChange
	}
}

func BackupDir(backupDir string) Option {
	return func(options *Options) {
		options.backupDir = backupDir
	}
}

func DebugOpt(debug bool) Option {
	return func(options *Options) {
		options.debug = debug
	}
}

func ClientIP(ip string) Option {
	return func(options *Options) {
		options.clientIP = ip
	}
}
