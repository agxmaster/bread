package rest

import (
	"context"
	"net/http"

	"git.qutoutiao.net/gopher/qms/internal/client/rest"
	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/invoker"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
)

const (
	//HTTP is url schema name
	HTTP  = "http"
	HTTPS = "https"
)

type Invoker struct {
	opt   *option
	dopts []DialOption
}

func NewInvoker(dopts ...DialOption) *Invoker {
	invoke := &Invoker{
		opt:   &option{},
		dopts: dopts,
	}

	for _, o := range dopts {
		o(invoke.opt)
	}

	return invoke
}

func (r *Invoker) Invoke(ctx context.Context, req *http.Request, opts ...CallOption) (*http.Response, error) {
	common.SetXQMSContext(map[string]string{common.HeaderSourceName: runtime.ServiceName}, req)

	// call option 每次都需要apply
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}
	if opt.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.timeout)
		defer cancel()
	}

	resp := rest.NewResponse()
	inv := r.request2Invocation(ctx, req, resp, opt)
	// invoke
	err := invoker.NewInvoker(r.dopts...).Invoke(inv, opts...)
	if err == client.ErrCanceled && ctx.Err() == context.DeadlineExceeded {
		err = client.ErrTimeout
	}
	return resp, err
}

func (r *Invoker) request2Invocation(ctx context.Context, req *http.Request, resp *http.Response, opt *option) *invocation.Invocation {
	inv := invocation.New(ctx, req.Host)

	// add headers to req
	for k, vv := range opt.header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// set headers to Ctx
	h := common.FromContext(inv.Ctx)
	for k, vv := range req.Header {
		h.Set(k, vv...)
	}

	_, port, _ := util.ParseServiceAndPort(req.Host)
	inv.Protocol = protocol.ProtocHTTP
	inv.Port = port
	inv.RouteType = opt.routeType
	inv.SchemaID = common.ProtocolRest
	inv.OperationID = req.URL.Path
	inv.Args = req
	inv.Reply = resp
	inv.URLPathFormat = req.URL.Path
	inv.BriefURI = req.URL.Path
	inv.Strategy = r.opt.balancerName
	inv.SetMetadata(common.RestMethod, req.Method)

	return inv
}
