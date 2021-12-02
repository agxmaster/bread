package invocation

import (
	"context"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/registryutil"
	utiltags "git.qutoutiao.net/gopher/qms/internal/pkg/util/tags"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// constant values for consumer and provider
const (
	Consumer = iota
	Provider
)

// Response is invocation response struct
type Response struct {
	Status    int
	Result    interface{}
	KVs       map[string]interface{}
	RequestID string
	Err       error
}

// ResponseCallBack process invocation response
type ResponseCallBack func(*Response) error

//Invocation is the basic struct that used in go sdk to make client and transport layer transparent .
//developer should implements a client which is able to transfer invocation to there own request
//a protocol server should transfer request to invocation and then back to request
type Invocation struct {
	HandlerIndex       int
	Endpoint           string //service's ip and port, it is decided in load balancing
	Protocol           protocol.Protocol
	Env                qenv.Env
	Port               string //Port is the name of a real service port
	SourceServiceID    string
	SourceMicroService string
	MicroServiceName   string //Target micro service name
	SchemaID           string //correspond struct name
	OperationID        string //correspond struct func name
	Args               interface{}
	URLPathFormat      string // 真实的URI
	BriefURI           string // 防止metrics膨胀
	Reply              interface{}
	Ctx                context.Context        //ctx can save protocol headers
	Metadata           map[string]interface{} //local scope data
	RouteTags          utiltags.Tags          //route tags is decided in router handler
	Strategy           string                 //load balancing strategy
	Filters            []string
	NoDiscovery        bool //if do not using service discovery
	IsStream           bool // is turn on stream mode
	StreamDesc         interface{}
	RouteType          string
	DialOptions        []base.OptionFunc
	CallOptions        []base.OptionFunc
	// grpc stream desc pointer
}

//Reset reset clear a invocation
func (inv *Invocation) Reset() {
	inv.Endpoint = ""
	inv.Protocol = protocol.ProtocUnknown
	inv.Env = qenv.Get()
	inv.SourceServiceID = ""
	inv.SourceMicroService = ""
	inv.MicroServiceName = ""
	inv.SchemaID = ""
	inv.OperationID = ""
	inv.Args = nil
	inv.URLPathFormat = ""
	inv.BriefURI = ""
	inv.Reply = nil
	inv.Ctx = nil
	inv.Metadata = nil
	inv.RouteTags = utiltags.Tags{}
	inv.Filters = nil
	inv.Strategy = ""
	inv.NoDiscovery = false
}

// New create invocation, context can not be nil
// if you don't set ContextHeaderKey, then New will init it
func New(ctx context.Context, remote string) *Invocation {
	inv := &Invocation{
		SourceServiceID:  runtime.ServiceID,
		MicroServiceName: remote,
		Ctx:              ctx,
	}
	if inv.Ctx == nil {
		inv.Ctx = context.TODO()
	}
	//create new map for ContextHeaderKey
	inv.Ctx = common.NewContext(inv.Ctx, common.Header{})
	inv.Env = inv.GetUpstream().Env

	return inv
}

//SetMetadata local scope params
func (inv *Invocation) SetMetadata(key string, value interface{}) {
	if inv.Metadata == nil {
		inv.Metadata = make(map[string]interface{})
	}
	inv.Metadata[key] = value
}

//SetHeader set headers, the client and server plugins should use them in protocol headers
//it is convenience but has lower performance than you use Headers[k]=v,
// when you have a batch of kv to set
func (inv *Invocation) SetHeader(k, v string) {
	h := common.FromContext(inv.Ctx)
	h.Set(k, v)
}

//Headers return a map that protocol plugin should deliver in transport
func (inv *Invocation) Headers() common.Header {
	return common.FromContext(inv.Ctx)
}

//Headers return a map that protocol plugin should deliver in transport
func (inv *Invocation) GetUpstream() *config.Upstream {
	return config.GetUpstream(inv.MicroServiceName)
}

func (inv *Invocation) GetRemoteService() string {
	return config.GetRemoteService(inv.MicroServiceName)
}

func (inv *Invocation) GetMeshService() (meshService string) {
	if meshService = inv.GetUpstream().Sidecar.MeshService; meshService == "" {
		meshService = registryutil.ToStdRegistryName(inv.Endpoint, inv.Protocol, inv.Env)
	}
	return
}
