package server

import (
	"fmt"
	"net"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	qmsTLS "git.qutoutiao.net/gopher/qms/internal/core/tls"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var (
	servers       = make(map[string]Server)
	serverPlugins = make(map[string]NewFunc)
)

//NewFunc returns a ProtocolServer
type NewFunc func(*InitOptions) Server

//InstallPlugin For developer
func InstallPlugin(protocol string, newFunc NewFunc) {
	serverPlugins[protocol] = newFunc
	qlog.Trace("Installed Server Plugin, protocol:" + protocol)
}

//GetServerFunc returns the server function
func GetServerFunc(protocol string) (NewFunc, error) {
	f, ok := serverPlugins[protocol]
	if !ok {
		return nil, fmt.Errorf("unknown protocol server [%s]", protocol)
	}
	return f, nil
}

//GetServer return the server based on protocol
func GetServer(protocol string) (Server, error) {
	s, ok := servers[protocol]
	if !ok {
		return nil, fmt.Errorf("[%s] server isn't running, please makesure it's configured in app.yaml ", protocol)
	}
	return s, nil
}

//GetServers returns the map of servers
func GetServers() map[string]Server {
	return servers
}

//ErrRuntime is an error channel, if it receive any signal will cause graceful shutdown of go chassis, process will exit
var ErrRuntime = make(chan error)

//StartServer starting the server
func StartServer() error {
	for name, server := range servers {
		qlog.Info("starting " + name + " server ...")
		err := server.Start()
		if err != nil {
			qlog.Errorf("servers failed to start, err %s", err)
			return fmt.Errorf("can not start [%s] server,%s", name, err.Error())
		}
		qlog.Info(name + " server start success")
	}
	qlog.Info("all server starting is completed")

	return nil
}

//UnRegistrySelfInstances this function removes the self instance
//func UnRegistrySelfInstances(action actionutil.Action, graceful config.Graceful) error {
//	services := make([]string, 0)
//	if action == actionutil.ActionClose {
//		for name := range config.GetServiceMap() {
//			services = append(services, name)
//		}
//	} else if action == actionutil.ActionReload {
//		services = append(services, graceful.Deregisters...)
//	}
//	if err := registry.DefaultRegistrator.UnRegisterMicroServiceInstance(services); err != nil {
//		qlog.Errorf("StartServer() UnregisterMicroServiceInstance failed, sid/iid: %s/%s: %s",
//			runtime.ServiceID, runtime.InstanceID, err)
//		return err
//	}
//	return nil
//}

//Init initializes
func Init(opts ...base.OptionFunc) error {
	opt := &InitOptions{
		listenM: make(map[string]net.Listener),
		Opts:    opts,
	}
	for _, o := range opts {
		o(opt)
	}

	qlog.Tracef("listens: %v", opt.listenM)
	var err error
	for name, spec := range config.GetServiceMap() {
		if err = initialServer(name, spec, opt); err != nil {
			qlog.Error(err)
			return err
		}
	}
	return nil

}

func initialServer(name string, spec config.ServiceSpec, opt *InitOptions) error {
	protocolName, _, err := util.ParsePortName(name)
	if err != nil {
		return err
	}
	qlog.Tracef("Init server [%s], protocol is [%s]", name, protocolName)
	f, err := GetServerFunc(protocolName)
	if err != nil {
		return fmt.Errorf("do not support [%s] server", name)
	}

	sslTag := name + "." + common.Provider
	listen := opt.listenM[spec.Address]
	tlsConfig, sslConfig, err := qmsTLS.GetTLSConfigByService("", name, common.Provider)
	if err != nil {
		if !qmsTLS.IsSSLConfigNotExist(err) {
			return err
		}
	} else {
		if listen != nil {
			return fmt.Errorf("grace restart and tls are not supported at the same time")
		}
		qlog.Warnf("%s TLS mode, verify peer: %t, cipher plugin: %s.",
			sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin)
	}

	if spec.Address == "" {
		spec.Address = iputil.DefaultEndpoint4Protocol(name)
	}

	s := f(&InitOptions{
		ServerName:    name,
		Address:       spec.Address,
		Listen:        listen,
		TLSConfig:     tlsConfig,
		EnableGrpcurl: spec.GrpcurlEnabled,
		Opts:          opt.Opts,
	})

	servers[name] = s
	return nil
}
