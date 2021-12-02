package grpc

import (
	"context"

	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/invoker"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
)

type Invoker struct {
	opt   *option
	dopts []client.DialOption
}

func NewInvoker(target string, dopts ...DialOption) *Invoker {
	invoke := &Invoker{
		opt: &option{
			remoteService: target,
		},
		dopts: dopts,
	}

	for _, o := range dopts {
		o(invoke.opt)
	}

	return invoke
}

func (g *Invoker) Invoke(ctx context.Context, operationID string, arg interface{}, reply interface{}, copts ...CallOption) error {
	for _, o := range copts {
		o(g.opt)
	}
	if g.opt.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.opt.timeout)
		defer cancel()
	}

	inv := invocation.New(ctx, g.opt.remoteService)
	inv.Protocol = protocol.ProtocGrpc
	inv.OperationID = operationID
	inv.Reply = reply
	inv.Args = arg
	inv.Strategy = g.opt.balancerName
	inv.RouteType = g.opt.routeType
	return invoker.NewInvoker(g.dopts...).Invoke(inv, copts...)
}
