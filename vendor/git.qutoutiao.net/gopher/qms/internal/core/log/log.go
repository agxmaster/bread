package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

// constant values for logrotate parameters
const (
	LogRotateDate     = 1
	LogRotateSize     = 10
	LogBackupCount    = 7
	RollingPolicySize = "size"
)

// logFilePath log file path
var logFilePath string

//Options is the struct for lager information(lager.yaml)
type Options struct {
	Output              string `yaml:"output"`
	LoggerLevel         string `yaml:"logger_level"`
	LoggerFile          string `yaml:"logger_file"`
	LogFormatText       bool   `yaml:"log_format_text"`
	EnableHTMLEscape    bool   `yaml:"enable_html_escape"`
	DisableReportCaller bool   `yaml:"disable_report_caller"`
	RollingPolicy       string `yaml:"rollingPolicy"`
	LogRotateDate       int    `yaml:"log_rotate_date"`
	LogRotateSize       int    `yaml:"log_rotate_size"`
	LogBackupCount      int    `yaml:"log_backup_count"`
}

// Init Build constructs a *Lager.Logger with the configured parameters.
func Init(option *Options) {
	logger := newLog(option)
	qlog.SetLogger(logger)

	if logFilePath != "" {
		initLogRotate(logFilePath, option)
	}
	logger.Debug("logger init success")
}

// newLog new log
func newLog(option *Options) qlog.Logger {
	var (
		output       = io.Writer(os.Stdout)
		formatter    = qlog.JsonFormatter
		reportCaller = true
		loggerLevel  = qlog.InfoLevel
	)

	if option.LoggerLevel != "" {
		level, err := qlog.ParseLevel(option.LoggerLevel)
		if err == nil {
			loggerLevel = level
		}
	}

	if option.DisableReportCaller {
		reportCaller = false
	}

	if option.Output == "file" {
		if option.LoggerFile == "" {
			option.LoggerFile = "log/qms.log"
		}
		if filepath.IsAbs(option.LoggerFile) {
			createLogFile("", option.LoggerFile)
			logFilePath = filepath.Join("", option.LoggerFile)
		} else {
			createLogFile(os.Getenv("QMS_HOME"), option.LoggerFile)
			logFilePath = filepath.Join(os.Getenv("QMS_HOME"), option.LoggerFile)
		}
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		output = file
	}

	if option.LogFormatText {
		formatter = qlog.TextFormatter
	}

	return qlog.NewWithOption(&qlog.Option{
		Output:           output,
		Level:            loggerLevel,
		Formatter:        formatter,
		EnableHTMLEscape: option.EnableHTMLEscape,
		ReportCaller:     reportCaller,
	})
}

// createLogFile create log file
func createLogFile(localPath, out string) {
	_, err := os.Stat(strings.Replace(filepath.Dir(filepath.Join(localPath, out)), "\\", "/", -1))
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(strings.Replace(filepath.Dir(filepath.Join(localPath, out)), "\\", "/", -1), os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(strings.Replace(filepath.Join(localPath, out), "\\", "/", -1), os.O_CREATE, 0640)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}
