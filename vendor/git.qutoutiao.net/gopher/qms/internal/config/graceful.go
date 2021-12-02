package config

import (
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
)

// 获取graceful需要的配置

type Graceful struct {
	Deregisters         []string // 需要去反注册的服务名称
	AutometricsDisabled bool     // 是否开启自动注册
}

func GracefulConfig() (Graceful, error) {
	graceful := Graceful{
		Deregisters: make([]string, 0),
	}

	graceconf := qconf.New()
	graceconf.AddOptionalFile(fileutil.AppConfigPath(), fileutil.AdvancedConfigPath())
	if err := graceconf.ReadInConfig(); err != nil { // read in memory
		return graceful, errors.WithStack(err)
	}

	registry := &Registry{}
	registry.init(graceconf)
	defaultRegisterDisabled := graceconf.GetBool(getServiceKey(constutil.Common, "registrator.disabled"), graceconf.GetBool("qms.service.registry.registerDisabled"))
	// 只关心存在的service的变更
	for service := range GetServiceMap() {
		if !registry.Disabled {
			if !graceconf.GetBool(getServiceKey(service, "registrator.disabled"), defaultRegisterDisabled) {
				continue
			}
		}
		graceful.Deregisters = append(graceful.Deregisters, service)
	}

	autoMetrics := &AutoMetrics{}
	autoMetrics.init(graceconf)
	graceful.AutometricsDisabled = !autoMetrics.Enabled

	return graceful, nil
}
