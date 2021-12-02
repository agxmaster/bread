package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/config/model"
	qmsTLS "git.qutoutiao.net/gopher/qms/internal/core/tls"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var (
	//ErrClientNotExist happens if client do not exist
	ErrClientNotExist = errors.New("client not exists")

	clients = sync.Map{}
)

//DefaultPoolSize is 500
const DefaultPoolSize = 512

//Options is configs for client creation
type Options struct {
	Service       string
	PoolSize      int
	Timeout       time.Duration
	RespCache     time.Duration
	RespCacheSize int
	Endpoint      string
	PoolTTL       time.Duration
	TLSConfig     *tls.Config
	Failure       map[string]bool
}

// GetFailureMap return failure map
func GetFailureMap(p string) map[string]bool {
	failureList := strings.Split(coreconf.GlobalDefinition.Qms.Transport.Failure[p], ",")
	failureMap := make(map[string]bool)
	for _, v := range failureList {
		if v == "" {
			continue
		}
		failureMap[v] = true
	}
	return failureMap
}

//GetMaxIdleCon get max idle connection number you defined
//default is 512
func GetMaxIdleCon(service string) int {
	return config.GetUpstream(service).Transport.MaxIdleConn
}

// CreateClient is for to create client based on protocol and the service name
func NewClient(protocol, service, endpoint string, dopts ...DialOption) (Client, error) {
	f, err := GetClientNewFunc(protocol)
	if err != nil {
		qlog.Error(fmt.Sprintf("do not support [%s] client", protocol))
		return nil, err
	}

	tlsConfig, sslConfig, err := qmsTLS.GetTLSConfigByService(service, protocol, common.Consumer)
	if err != nil {
		if !qmsTLS.IsSSLConfigNotExist(err) {
			return nil, err
		}
	} else {
		// client verify target micro service's name in mutual tls
		// remember to set SAN (Subject Alternative Name) as server's micro service name
		// when generating server.csr
		tlsConfig.ServerName = service
		qlog.Warnf("%s %s TLS mode, verify peer: %t, cipher plugin: %s.",
			protocol, service, sslConfig.VerifyPeer, sslConfig.CipherPlugin)
	}
	return f(Options{
		Service:       service,
		TLSConfig:     tlsConfig,
		PoolSize:      GetMaxIdleCon(service),
		Failure:       GetFailureMap(protocol),
		Timeout:       config.GetTimeoutDuration(service),
		RespCache:     config.GetRespCacheDuration(service),
		RespCacheSize: config.GetRespCacheSize(service),
		Endpoint:      endpoint,
	}, dopts...)
}

func generateKey(protocol, service, endpoint string) string {
	return protocol + service + endpoint
}

// GetClient is to get the client based on protocol, service,endpoint name
func GetClient(protocol, service, endpoint string, opts ...DialOption) (Client, error) {
	key := generateKey(protocol, service, endpoint)

	value, ok := clients.Load(key)
	if !ok {
		qlog.Infof("create new client for protocol=%s, service=%s, endpoint=%s", protocol, service, endpoint)

		client, err := NewClient(protocol, service, endpoint, opts...)
		if err != nil {
			qlog.Errorf("create new client for protocol=%s, service=%s, endpoint=%s: %+v", protocol, service, endpoint, err)

			return nil, err
		}

		clients.Store(key, client)

		return client, nil
	}

	client, ok := value.(Client)
	if !ok {
		qlog.Warn("invalid client(%T) for protocol=%s, service=%s, endpoint=%s", client, protocol, service, endpoint)

		client, err := NewClient(protocol, service, endpoint, opts...)
		if err != nil {
			qlog.Errorf("create new client for protocol=%s, service=%s, endpoint=%s: %+v", protocol, service, endpoint, err)

			return nil, err
		}

		clients.Store(key, client)

		return client, nil
	}

	return client, nil
}

//Close close a client conn
func Close(protocol, service, endpoint string) error {
	key := generateKey(protocol, service, endpoint)

	value, ok := clients.Load(key)
	if !ok {
		return nil
	}
	clients.Delete(key)

	client, ok := value.(Client)
	if !ok {
		return nil
	}

	err := client.Close()
	if err != nil {
		qlog.Errorf("close client for protocol=%s, service=%s, endpoint=%s: %+v", protocol, service, endpoint, err)
	}

	return err
}

// SetTimeoutToClientCache set timeout to client
func SetTimeoutToClientCache(spec *model.IsolationWrapper) {
	clients.Range(func(key, value interface{}) bool {
		client, ok := value.(Client)
		if ok {
			if v, ok := spec.Consumer.AnyService[client.GetOptions().Service]; ok {
				client.ReloadConfigs(Options{Timeout: time.Duration(v.TimeoutInMilliseconds) * time.Millisecond})
			} else {
				client.ReloadConfigs(Options{Timeout: time.Duration(spec.Consumer.TimeoutInMilliseconds) * time.Millisecond})
			}
		}
		return true
	})
}

// EqualOpts equal newOpts and oldOpts
func EqualOpts(oldOpts, newOpts Options) Options {
	if newOpts.Timeout != oldOpts.Timeout {
		oldOpts.Timeout = newOpts.Timeout
	}

	if newOpts.PoolSize != 0 {
		oldOpts.PoolSize = newOpts.PoolSize
	}
	if newOpts.PoolTTL != 0 {
		oldOpts.PoolTTL = newOpts.PoolTTL
	}
	if newOpts.TLSConfig != nil {
		oldOpts.TLSConfig = newOpts.TLSConfig
	}
	oldOpts.Failure = newOpts.Failure
	return oldOpts
}
