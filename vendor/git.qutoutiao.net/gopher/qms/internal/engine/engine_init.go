package engine

import (
	"fmt"
	"net"
	"sync"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/bootstrap"
	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/balancer"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/registry"
	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"git.qutoutiao.net/gopher/qms/internal/core/tracing"
	"git.qutoutiao.net/gopher/qms/internal/eventlistener"
	"git.qutoutiao.net/gopher/qms/internal/native"
	"git.qutoutiao.net/gopher/qms/internal/pkg/circuit"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qms/third_party/forked/afex/hystrix-go/hystrix"
)

type engine struct {
	version     string
	schemas     []*Schema
	mu          sync.Mutex
	Initialized bool
	native      *native.Native

	DefaultConsumerChainNames map[string]string
	DefaultProviderChainNames map[string]string
}

// Schema struct for to represent schema info
type Schema struct {
	option *serverOptions
	schema interface{}
	opts   []server.Option
}

func (c *engine) initChains(chainType string) error {
	var defaultChainName = "default"
	var handlerNameMap = map[string]string{defaultChainName: ""}
	switch chainType {
	case common.Provider:
		if providerChainMap := coreconf.GlobalDefinition.Qms.Handler.Chain.Provider; len(providerChainMap) != 0 {
			if _, ok := providerChainMap[defaultChainName]; !ok {
				providerChainMap[defaultChainName] = c.DefaultProviderChainNames[defaultChainName]
			}
			handlerNameMap = providerChainMap
		} else {
			handlerNameMap = c.DefaultProviderChainNames
		}
	case common.Consumer:
		if consumerChainMap := coreconf.GlobalDefinition.Qms.Handler.Chain.Consumer; len(consumerChainMap) != 0 {
			if _, ok := consumerChainMap[defaultChainName]; !ok {
				consumerChainMap[defaultChainName] = c.DefaultConsumerChainNames[defaultChainName]
			}
			handlerNameMap = consumerChainMap
		} else {
			handlerNameMap = c.DefaultConsumerChainNames
		}
	}
	qlog.Tracef("init %s's handler map", chainType)
	return handler.CreateChains(chainType, handlerNameMap)
}

func (c *engine) initHandler() error {
	if err := c.initChains(common.Provider); err != nil {
		qlog.Errorf("chain int failed: %s", err)
		return err
	}
	if err := c.initChains(common.Consumer); err != nil {
		qlog.Errorf("chain int failed: %s", err)
		return err
	}
	qlog.Trace("chain init success")
	return nil
}

//Init
func (c *engine) initialize(options ...base.OptionFunc) error {
	if c.Initialized {
		return nil
	}

	var opts initOptions
	for _, o := range options {
		o(&opts)
	}

	if opts.configDir != "" {
		qlog.Infof("set config dir to %s", opts.configDir)
		fileutil.SetConfDir(opts.configDir)
	}

	if err := config.Init(); err != nil {
		qlog.Error("failed to initialize conf: " + err.Error())
		return err
	}

	err := c.initHandler()
	if err != nil {
		qlog.Errorf("handler init failed: %s", err)
		return err
	}

	if err := metrics.Init(opts.metricLabels...); err != nil {
		return err
	}

	addrM := make(map[string]net.Listener)
	if overss.Enabled {
		for i := 0; i < len(overss.Addresses); i++ {
			options = append(options, server.WithListener(overss.Addresses[i], overss.Listeners[i]))
			addrM[overss.Addresses[i]] = overss.Listeners[i]
		}
	}
	// 初始化native
	c.native = native.NewNative()
	c.native.Init(addrM)

	err = server.Init(options...)
	if err != nil {
		return err
	}
	bootstrap.Bootstrap()
	if !config.Get().Registry.Disabled {
		if err := registry.Init(); err != nil {
			return errors.WithStack(err)
		}

		// 开启LB 要在control.Init之前设置，因为panel会加缓存
		// 不严谨 自定义服务路由 也应该开启该配置的
		balancer.Enable()
	}

	ctlOpts := control.Options{
		Infra:   coreconf.GlobalDefinition.Panel.Infra,
		Address: coreconf.GlobalDefinition.Panel.Settings["address"],
	}
	if err := control.Init(ctlOpts); err != nil {
		return err
	}

	if !config.Get().Trace.Disabled {
		if err = tracing.Init(); err != nil {
			return err
		}
	}
	go hystrix.StartReporter()
	circuit.Init()
	eventlistener.Init()
	c.Initialized = true
	return nil
}

func (c *engine) registerSchema(defaultServerName string, structPtr interface{}, opts ...ServerOption) {
	schema := &Schema{
		option: &serverOptions{
			serverName: defaultServerName,
		},
		schema: structPtr,
		opts:   opts,
	}

	for _, o := range opts {
		o(schema.option)
	}

	c.mu.Lock()
	c.schemas = append(c.schemas, schema)
	c.mu.Unlock()
}

func (c *engine) start() error {
	if !c.Initialized {
		return fmt.Errorf("the qms do not init. please run qms.Init() first")
	}

	// start native
	go func() {
		if err := c.native.Run(); err != nil {
			qlog.Warn("native server err: " + err.Error())
			server.ErrRuntime <- err
		}
	}()

	for _, v := range c.schemas {
		if v == nil {
			continue
		}
		s, err := server.GetServer(v.option.serverName)
		if err != nil {
			return err
		}
		_, err = s.Register(v.schema, v.opts...)
		if err != nil {
			return err
		}
	}
	err := server.StartServer()
	if err != nil {
		return err
	}

	// program ready[要放到最后]
	overss.ProgramReady()

	return nil
}
