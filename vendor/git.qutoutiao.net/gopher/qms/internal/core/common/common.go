package common

import (
	"context"
	"net/http"

	"git.qutoutiao.net/gopher/qms/pkg/json"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

// constant for provider and consumer
const (
	Provider = "Provider"
	Consumer = "Consumer"
)

const (
	// ScopeFull means service is able to access to another app's service
	ScopeFull = "full"
	// ScopeApp means service is not able to access to another app's service
	ScopeApp = "app"
)

// constant for micro service environment parameters
const (
	Env = "go-chassis_ENV"

	EnvNodeIP      = "HOSTING_SERVER_IP"
	EnvSchemaRoot  = "SCHEMA_ROOT"
	EnvProjectID   = "QMS_PROJECT_ID"
	EnvQMSEndpoint = "PAAS_QMS_ENDPOINT"
)

// constant environment keys service center, config center, monitor server addresses
const (
	CseRegistryAddress     = "QMS_REGISTRY_ADDR"
	CseConfigCenterAddress = "QMS_CONFIG_CENTER_ADDR"
	CseMonitorServer       = "QMS_MONITOR_SERVER_ADDR"
	EnvQMSSCEndpoint       = "PAAS_QMS_SC_ENDPOINT"
	EnvQMSCCEndpoint       = "PAAS_QMS_CC_ENDPOINT"
)

// env connect with "." like service_description.name and service_description.version which can not be used in k8s.
// So we can not use archaius to set env.
// To support this declaring constant for service name and version
// constant for service name and version.
const (
	ServiceName = "SERVICE_NAME"
	Version     = "VERSION"
)

// constant for microservice environment
const (
	EnvValueDev  = "development"
	EnvValueProd = "production"
)

// constant for secure socket layer parameters
const (
	SslCipherPluginKey = "cipherPlugin"
	SslVerifyPeerKey   = "verifyPeer"
	SslCipherSuitsKey  = "cipherSuits"
	SslProtocolKey     = "protocol"
	SslCaFileKey       = "caFile"
	SslCertFileKey     = "certFile"
	SslKeyFileKey      = "keyFile"
	SslCertPwdFileKey  = "certPwdFile"
	AKSKCustomCipher   = "qms.credentials.akskCustomCipher"
)

// constant for protocol types
const (
	ProtocolRest    = "rest"
	ProtocolHighway = "highway"
	LBSessionID     = "go-chassisLB"
	ProtocolGrpc    = "grpc"
)

// configuration placeholders
const (
	PlaceholderInternalIP = "$INTERNAL_IP"
)

// SessionNameSpaceKey metadata session namespace key
const SessionNameSpaceKey = "_Session_Namespace"

// SessionNameSpaceDefaultValue default session namespace value
const SessionNameSpaceDefaultValue = "default"

// DefaultKey default key
const DefaultKey = "default"

// DefaultValue default value
const DefaultValue = "default"

// BuildinTagApp build tag for the application
const BuildinTagApp = "app"

// BuildinTagVersion build tag version
const BuildinTagVersion = "version"

// BuildinLabelVersion build label for version
const BuildinLabelVersion = BuildinTagVersion + ":" + LatestVersion

// CallerKey caller key
const CallerKey = "caller"

const (
	// HeaderSourceName is constant for header source name
	HeaderSourceName = "x-qms-src-microservice"
	// HeaderXCseContent is constant for header , get some json msg about HeaderSourceName like {"k":"v"}
	HeaderXCseContent = "x-qms-context"
	// HeaderSourceName is constant for header service name for sidecar
	HeaderXSidecar = "X-Qtt-Meshservice"
)

const (
	// SidecarAddress is constant for sidecar agent address.
	SidecarAddress = "127.0.0.1:8102"
	PilotAddress   = "pilot.1sapp.com:80"
)

// constant string for route type
const (
	RouteDiscovery = "discovery" //指定采用服务发现
	RouteDirect    = "direct"    //指定采用SLB或ipport
	RouteSidecar   = "sidecar"   //指定采用sidecar
	RouteDefault   = ""          //自动判断。(如果url.Host含有"."，则判为SLB/ipport，否则，判为服务发现)
)

// constant string for API path.
const (
	DefaultHealthzPath = "/ping"    //默认的健康检测接口
	DefaultMetricsPath = "/metrics" //默认的metrics接口
)

const (
	// RestMethod is the http method for restful protocol
	RestMethod = "method"
)

// constant for default application name and version
const (
	DefaultApp        = "default"
	DefaultVersion    = "v0.0.1"
	LatestVersion     = "latest"
	AllVersion        = "0+"
	DefaultStatus     = "UP"
	TESTINGStatus     = "TESTING"
	DefaultLevel      = "BACK"
	DefaultHBInterval = 30
)

//constant used
const (
	HTTP   = "http"
	HTTPS  = "https"
	JSON   = "application/json"
	Create = "CREATE"
	Update = "UPDATE"
	Delete = "DELETE"

	Client           = "client"
	File             = "File"
	DefaultTenant    = "default"
	DefaultChainName = "default"

	FileRegistry      = "File"
	DefaultUserName   = "default"
	DefaultDomainName = "default"
	DefaultProvider   = "default"

	TRUE  = "true"
	FALSE = "false"
)

// const default config for config-center
const (
	DefaultRefreshMode = 1
)

//ContextHeaderKey is the unified key of header value in context
//all protocol integrated with go chassis must set protocol header into context in this context key
type ContextHeaderKey struct{}

// NewContext transforms a metadata to context object
func NewContext(ctx context.Context, h Header) context.Context {
	if h == nil {
		return context.WithValue(ctx, ContextHeaderKey{}, Header{})
	}
	return context.WithValue(ctx, ContextHeaderKey{}, h)
}

// WithContext sets the KV and returns the context object
func WithContext(ctx context.Context, key, val string) context.Context {
	if ctx == nil {
		return NewContext(context.Background(), Header{
			key: []string{val},
		})
	}

	h := FromContext(ctx)
	h.Set(key, val)
	return ctx
}

// FromContext return the headers which should be send to provider
// through transport
func FromContext(ctx context.Context) Header {
	if ctx == nil {
		return Header{}
	}
	at, ok := ctx.Value(ContextHeaderKey{}).(Header)
	if !ok {
		return Header{}
	}
	return at
}

// GetXQMSContext  get x-qms-context from req.header
func GetXQMSContext(k string, r *http.Request) string {
	if r == nil || r.Header == nil {
		qlog.Trace("get x-qms-header failed , request(request.Header) is nil or  key is empty, please check its")
		return ""
	}
	cseContextStr := r.Header.Get(HeaderXCseContent)
	if cseContextStr == "" {
		return r.Header.Get(k)
	}

	var m map[string]string
	err := json.Unmarshal([]byte(cseContextStr), &m)
	if err != nil {
		qlog.Tracef("get x-qms-header form req failed , error : %v", err)
		return ""
	}
	return m[k]
}

// SetXQMSContext  set value into x-qms-context
func SetXQMSContext(kvs map[string]string, r *http.Request) {
	if len(kvs) <= 0 || r == nil {
		qlog.Trace("set x-qms-header into req failed, because one of kvs or request is empty(nil) or all are empty(nil)")
		return
	}

	if r.Header == nil {
		r.Header = http.Header{}
	}

	b, err := json.Marshal(kvs)
	if err != nil {
		qlog.Tracef("set value to x-qms-context failed, json.Marshal(%+v): %+v", kvs, err)
		return
	}

	r.Header.Set(HeaderXCseContent, string(b))
}
