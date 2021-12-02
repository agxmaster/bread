package metrics

import (
	"fmt"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

func init() {
	registries["prometheus"] = NewPrometheusExporter
}

//Init load the metrics plugin and initialize it
func Init(labels ...CustomLabel) error {
	name := "prometheus"

	f, ok := registries[name]
	if !ok {
		return fmt.Errorf("can not init metrics registry [%s]", name)
	}

	defaultRegistry = f(Options{
		FlushInterval: 10 * time.Second,
	})

	if err := enableProviderMetrics(labels...); err != nil {
		return err
	}
	if err := enableConsumerMetrics(); err != nil {
		return err
	}

	metrics := config.Get().Metrics
	if !runtime.InsideDocker && metrics.AutoMetrics.Enabled && !isRegistryInstance() {
		if err := enableAutoRegistryMetrics(); err != nil {
			qlog.Error(err)
		}
	}

	if !metrics.RedisDisabled {
		if err := disableRedisMetrics(); err != nil {
			qlog.Errorf("init redis metrics failed, err:", err.Error())
		}
	}

	if !metrics.GormDisabled {
		if err := disableGormMetrics(); err != nil {
			qlog.Errorf("init gorm metrics failed, err:", err.Error())
		}
	}

	if !metrics.ErrorLogDisabled {
		if err := enableErrorLogMetrics(); err == nil {
			qlog.SetMetricsFunc(countErrorNumber)
		} else {
			qlog.Error(err)
		}
	}

	return nil
}
