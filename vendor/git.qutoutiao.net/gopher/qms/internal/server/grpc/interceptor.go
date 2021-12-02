package grpc

import (
	"context"
	"fmt"
	"runtime"

	"git.qutoutiao.net/gopher/qms/internal/core/requestid"

	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func wrapUnaryInterceptor(chainName, serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, info.FullMethod)
			}
		}()

		c, err := handler.GetChain(common.Provider, chainName)
		if err != nil {
			qlog.Error(fmt.Sprintf("Handler chain init err [%s]", err.Error()))
			return nil, err
		}
		inv := Request2Invocation(ctx, serviceName, req, info)
		var r *invocation.Response
		c.Next(inv, func(ir *invocation.Response) error {
			defer func() {
				r = ir
			}()

			if ir.Err != nil {
				return ir.Err
			}

			ir.Result, ir.Err = handle(inv.Ctx, req)
			ir.Status = int(status.Code(ir.Err))
			ir.RequestID = requestid.FromContext(inv.Ctx)
			return ir.Err
		})
		return r.Result, r.Err
	}
}

func wrapStreamInterceptor(chainName, serviceName string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handle grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, info.FullMethod)
			}
		}()

		c, err := handler.GetChain(common.Provider, chainName)
		if err != nil {
			qlog.Error(fmt.Sprintf("Handler chain init err [%s]", err.Error()))
			return err
		}

		wrappedStream := grpc_middleware.WrapServerStream(stream)
		inv := Stream2Invocation(serviceName, wrappedStream, info)
		c.Next(inv, func(ir *invocation.Response) error {
			err = ir.Err
			if err != nil {
				return err
			}

			wrappedStream.WrappedContext = inv.Ctx
			err = handle(srv, wrappedStream)
			return err
		})
		return err
	}
}

func recoverFrom(r interface{}, fullMethod string) error {
	var stacktrace string
	for i := 1; ; i++ {
		_, f, l, got := runtime.Caller(i)
		if !got {
			break
		}

		stacktrace += fmt.Sprintf("%s:%d\n", f, l)
	}
	qlog.WithFields(qlog.Fields{
		"fullMethod": fullMethod,
		"panic":      r,
		"stack":      stacktrace,
	}).Error("handle request panic.")

	return status.Errorf(codes.Internal, "%s", r)
}
