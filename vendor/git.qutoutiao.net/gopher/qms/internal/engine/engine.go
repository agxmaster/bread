package engine

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	//init logger first
	"git.qutoutiao.net/gopher/qms/internal/initiator"
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"

	//load balancing
	_ "git.qutoutiao.net/gopher/qms/internal/balancer/random"
	_ "git.qutoutiao.net/gopher/qms/internal/balancer/roundrobin"
	_ "git.qutoutiao.net/gopher/qms/internal/balancer/weightedrandom"
	_ "git.qutoutiao.net/gopher/qms/internal/balancer/weightedroundrobin"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"

	//protocols
	_ "git.qutoutiao.net/gopher/qms/internal/client/grpc"
	_ "git.qutoutiao.net/gopher/qms/internal/client/rest"
	_ "git.qutoutiao.net/gopher/qms/internal/server/ginhttp"
	_ "git.qutoutiao.net/gopher/qms/internal/server/grpc"

	//routers
	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/registry"

	//router
	_ "git.qutoutiao.net/gopher/qms/internal/core/router/servicecomb"
	//control panel
	_ "git.qutoutiao.net/gopher/qms/internal/control/archaius"
	// registry
	"git.qutoutiao.net/gopher/qms/internal/core/metadata"
	_ "git.qutoutiao.net/gopher/qms/internal/core/registry/consul"
	"git.qutoutiao.net/gopher/qms/internal/core/server"

	//trace
	_ "git.qutoutiao.net/gopher/qms/internal/core/tracing/jaeger"
	// prometheus reporter for circuit breaker metrics
	_ "git.qutoutiao.net/gopher/qms/third_party/forked/afex/hystrix-go/hystrix/reporter"
	// aes package handles security related plugins
	_ "git.qutoutiao.net/gopher/qms/internal/security/plugins/aes"
	_ "git.qutoutiao.net/gopher/qms/internal/security/plugins/plain"

	//set GOMAXPROCS
	_ "go.uber.org/automaxprocs"

	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/actionutil"
	"git.qutoutiao.net/gopher/qms/third_party/forked/jpillora/overseer"
)

var (
	egn    *engine
	overss overseer.State
)

func init() {
	egn = &engine{}
}

//RegisterSchema Register a API service to specific server by name.
func RegisterSchema(defaultServerName string, structPtr interface{}, opts ...ServerOption) {
	egn.registerSchema(defaultServerName, structPtr, opts...)
}

func GetServer(defaultServerName string, opts ...ServerOption) (server.Server, error) {
	if !egn.Initialized {
		return nil, fmt.Errorf("the qms do not init. please run qms.Init() first")
	}

	opt := &serverOptions{
		serverName: defaultServerName,
	}
	for _, o := range opts {
		o(opt)
	}

	return server.GetServer(opt.serverName)
}

//Run bring up the service, it waits for os signal,and shutdown gracefully.
//it support graceful restart default, you can disable it by using qms.DisableGracefulRestart() as options.
func Run(options ...RunOption) error {
	var opts runOptions
	for _, o := range options {
		o(&opts)
	}

	// TODO: run一个服务 会有before run，after run，既然遍历起服务，何不如写在一块。比如现在注册和反注册也需要自己判断哪些服务需要注册和反注册 蛋疼
	err := egn.start()
	if err != nil {
		qlog.Error("run engine failed:" + err.Error())
		return err
	}

	if !config.GetRegistratorDisable() {
		//Register instance after Server started
		if err := registry.RegisterServices(); err != nil {
			qlog.Error("register instance failed:" + err.Error())
			return err
		}
	}

	// upload_swagger_api
	//go swagger.UploadSwaggerApi()

	waitingSignal(opts)
	return nil
}

func waitingSignal(opts runOptions) {
	var (
		c              = make(chan os.Signal) //Graceful shutdown
		action         actionutil.Action
		gracefulConfig config.Graceful
		err            error
	)

	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)

	select {
	case s := <-c:
		qlog.Info("got os signal " + s.String())
		action = actionutil.ActionClose
	case <-overss.Graceful():
		qlog.Info("got graceful shutdown signal from overseer for restart")
		action = actionutil.ActionReload
		gracefulConfig, err = config.GracefulConfig() // 设置graceful需要的配置
		if err != nil {
			qlog.WithError(err).Error("graceful时，重新获取配置失败.")
		}
	case err := <-server.ErrRuntime:
		qlog.Info("got server error " + err.Error())
	}

	if !config.GetRegistratorDisable() {
		registry.DeregisterServices(action, gracefulConfig)
	}

	if !runtime.InsideDocker && config.GetMetricsEnabled() && (action == actionutil.ActionClose || gracefulConfig.AutometricsDisabled) {
		metrics.DeAutoRegistryMetrics()
	}

	if runtime.InsideDocker {
		const WaitTime = 5
		waitTime := qconf.GetInt("qms.graceful.stop.docker_wait_time", WaitTime)
		if waitTime < 0 {
			waitTime = WaitTime
		}
		qlog.Infof("[in docker]sleep %d seconds before graceful shutdown", waitTime)
		time.Sleep(time.Second * time.Duration(waitTime))
	}

	for name, s := range server.GetServers() {
		qlog.Info("stopping " + name + " server...")
		err := s.Stop()
		if err != nil {
			qlog.Warnf("servers failed to stop: %s", err)
		}
		qlog.Info(name + " server stop success")
	}

	if err := egn.native.Stop(); err != nil {
		qlog.Warnf("servers failed to stop: %s", err)
	}
	qlog.Info("native server stop success")

	if opts.exitCb != nil {
		opts.exitCb()
	}

	qlog.Info("qms server gracefully shutdown")
}

//Init prepare the qms framework runtime
func Init(options ...InitOption) error {
	if egn.DefaultConsumerChainNames == nil {
		defaultChain := strings.Join([]string{
			handler.MetricsConsumer,
			handler.RatelimiterConsumer,
			handler.BizkeeperConsumer,
			handler.Loadbalance,
			handler.TracingConsumer,
			handler.Transport,
		}, ",")
		egn.DefaultConsumerChainNames = map[string]string{
			common.DefaultKey: defaultChain,
		}
	}
	if egn.DefaultProviderChainNames == nil {
		defaultChain := strings.Join([]string{
			handler.MetricsProvider,
			handler.RatelimiterProvider,
			handler.LogProvider,
			handler.TracingProvider,
		}, ",")
		egn.DefaultProviderChainNames = map[string]string{
			common.DefaultKey: defaultChain,
		}
	}
	//baseOpts := make([]base.OptionFunc, len(options))
	//for _, o := range options {
	//	baseOpts = append(baseOpts, base.OptionFunc(o))
	//}
	if err := egn.initialize(options...); err != nil {
		qlog.Info("init qms fail:", err)
		return err
	}

	qlog.Infof("init qms success, version is %s", metadata.SdkVersion)
	return nil
}

//GraceFork supports graceful restart by master/slave processes. [before init]
func GraceFork(main func()) {
	graceful := initiator.GetGraceful()
	if !graceful.Enabled {
		qlog.Info("graceful restart is disabled")
		main()
		return
	}

	// 如果开启了 必须指定地址
	var addresses []string
	for _, value := range graceful.Services {
		addresses = append(addresses, value)
	}
	if len(addresses) == 0 {
		panic(fmt.Errorf("listen address not parsed from coreconf file"))
	}
	// 添加native服务
	addresses = append(addresses, graceful.NativeAddress)
	sort.Strings(addresses)

	prog := func(state overseer.State) {
		qlog.Tracef("got overseer state: %+v", state)
		overss = state
		main()
	}

	overseer.Run(overseer.Config{
		Program:          prog,
		Addresses:        addresses,
		RestartPort:      graceful.ReloadPort,
		TerminateTimeout: time.Second * 30,
		RestartTimeout:   time.Duration(graceful.ReloadTimeoutMs) * time.Millisecond,
		Logger:           qlog.GetLogger(),
	})
}

func (e *engine) startServers() error {
	// 启动服务
	// 注册服务
	// 反注册服务
	return nil
}
