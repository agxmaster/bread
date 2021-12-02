// 框架相关逻辑，配合engine使用

package registry

import (
	"sync"

	corconf "git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/actionutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
)

// 默认只会有一个registry 不需要map
var (
	defaultRegistry Registry
	initOnce        sync.Once

	defaultBuilderName = "consul"
)

// Init 如果使能registry时，需要先进行初始化
func Init() (err error) {
	initOnce.Do(func() {
		// builder
		builder := getBuilder(defaultBuilderName)
		if builder == nil {
			err = errors.Wrapf(ErrNotFoundBuilder, "builder name: %s", defaultBuilderName)
			return
		}
		// config
		var conf *Config
		conf, err = GetConfig()
		if err != nil {
			err = errors.Wrapf(ErrGetConfig, "get config error: %v", err)
			return
		}
		// registry
		defaultRegistry, err = builder.Build(conf)
	})
	return
}

func GetRegistry() Registry {
	return defaultRegistry
}

// TODO: 临时放在此处
func RegisterServices() error {
	// 注册所有的services到consul
	for name, service := range corconf.Get().Service.ServiceM {
		if service.Registrator.Disabled || name == constutil.Common {
			continue
		}

		_, port, err := iputil.SplitHostPort(service.Address)
		if err != nil {
			return errors.WithStack(err)
		}

		if err = defaultRegistry.Register(NewService(port, service.Protocol)); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// TODO: 临时放在此处
func DeregisterServices(action actionutil.Action, graceful corconf.Graceful) {
	services := make([]string, 0)
	if action == actionutil.ActionClose {
		for name := range corconf.GetServiceMap() {
			services = append(services, name)
		}
	} else if action == actionutil.ActionReload {
		services = append(services, graceful.Deregisters...)
	}

	for _, service := range services {
		defaultRegistry.Deregister(protocol.ToProtocol(service))
	}
}
