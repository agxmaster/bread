//Package initiator init necessary module
// before every other package init functions
package initiator

import (
	"fmt"

	"git.qutoutiao.net/gopher/qms/internal/core/log"
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"github.com/spf13/cast"
)

// 在执行 qms.Init 前需要提前获取的信息

func init() {
	// logger and graceful config
	if err := initConfig(fileutil.AppConfigPath(), fileutil.AdvancedConfigPath(), fileutil.LogConfigPath()); err != nil {
		panic(err)
	}

	InitLogger()
}

func initConfig(files ...string) error {
	// 初始化配置文件[为了解耦配置和配置文件 框架需要的配置 必须放在下列三个文件下]
	qconf.AddOptionalFile(files...)
	if err := qconf.ReadInConfig(); err != nil { // read in memory
		return errors.WithStack(err)
	}
	return nil
}

// InitLogger initiate config file and openlogging before other modules
func InitLogger() {
	opts := &log.Options{
		LoggerLevel:      qconf.GetString("qms.logger.level", qconf.GetString("logger_level", defaultLogLevel)),
		Output:           qconf.GetString("qms.logger.output", qconf.GetString("output", defaultLogOutput)),
		EnableHTMLEscape: qconf.GetBool("qms.logger.html_escape_enabled", qconf.GetBool("enable_html_escape")),
		LoggerFile:       qconf.GetString("qms.logger.file", qconf.GetString("logger_file", defaultLogFile)),
		RollingPolicy:    qconf.GetString("qms.logger.rolling_policy", qconf.GetString("rollingPolicy", defaultLogRollingPolicy)),
		LogRotateDate:    qconf.GetInt("qms.logger.rotate_date", qconf.GetInt("log_rotate_date", defaultLogRotateDate)),
		LogRotateSize:    qconf.GetInt("qms.logger.rotate_size", qconf.GetInt("log_rotate_size", defaultLogRotateSize)),
		LogBackupCount:   qconf.GetInt("qms.logger.backup_count", qconf.GetInt("log_backup_count", defaultLogBackupCount)),
	}
	if qconf.GetString("qms.logger.format") == logFormatText || qconf.GetBool("log_format_text") {
		opts.LogFormatText = true
	}

	qlog.Debugf("log configs: %+v", opts)
	log.Init(opts)
}

type Graceful struct {
	Enabled         bool
	ReloadPort      int
	ReloadTimeoutMs int
	Services        map[string]string // map[name]address
	NativeAddress   string
}

func GetGraceful() Graceful {
	if enabled := qconf.GetBool("qms.graceful.enabled"); enabled {
		graceful := Graceful{
			Enabled:         true,
			ReloadPort:      qconf.GetInt("qms.graceful.reload_port", qconf.GetInt("qms.graceful.reloadPort", defaultReloadPort)),
			ReloadTimeoutMs: qconf.GetInt("qms.graceful.reload_timeout", defaultReloadTimeMs),
			Services:        make(map[string]string),
		}
		// 获取服务端口信息
		for service, value := range qconf.GetStringMap("qms.service", qconf.GetStringMap("qms.protocols")) {
			if service == constutil.Common {
				continue
			}
			if valueM, err := cast.ToStringMapE(value); err == nil {
				address, ok := valueM["address"]
				if !ok {
					address = valueM["listenaddress"] // 由于被qconf处理，所以key为小写
				}
				if address = cast.ToString(address); address != "" {
					graceful.Services[service] = cast.ToString(address)
				}
			}
		}
		nativePort := qconf.GetInt("qms.native.port", defaultNativePort)
		graceful.NativeAddress = fmt.Sprintf("0.0.0.0:%d", nativePort)
		return graceful
	}
	return Graceful{}

}
