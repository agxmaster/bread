package grpc

import (
	"context"

	clientgrpc "git.qutoutiao.net/gopher/qms/internal/client/grpc"
	invokergrpc "git.qutoutiao.net/gopher/qms/internal/invoker/grpc"
	"google.golang.org/grpc"
)

type Invoker struct {
	i *invokergrpc.Invoker
}

// NewInvoker 创建Invoker 类似client，由pb生成使用
func NewInvoker(target string, dopts ...DialOption) *Invoker {
	return &Invoker{
		i: invokergrpc.NewInvoker(target, dopts...),
	}
}

// Invoke 调用unary API
func (i *Invoker) Invoke(ctx context.Context, operationID string, arg interface{}, reply interface{}, options ...CallOption) error {
	return i.i.Invoke(ctx, operationID, arg, reply, options...)
}

// Stream 调用Stream API
func (i *Invoker) Stream(ctx context.Context, operationID string, streamDesc *grpc.StreamDesc, reply interface{}, options ...CallOption) error {
	options = append(options, clientgrpc.StreamDesc(streamDesc))
	return i.i.Invoke(ctx, operationID, nil, reply, options...)
}
