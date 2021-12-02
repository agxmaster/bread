package grpc

import (
	"git.qutoutiao.net/gopher/qms/internal/base"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type CallOption = base.OptionFunc

type callOption struct {
	streamDesc *grpc.StreamDesc
	opts       []grpc.CallOption // 起到过滤和拦截的作用，qms暴露的才能传递
}

// Header returns a CallOptions that retrieves the header metadata
// for a unary RPC.
func Header(md *metadata.MD) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*callOption); ok {
			o.opts = append(o.opts, grpc.Header(md))
		}
	}
}

// Trailer returns a CallOptions that retrieves the trailer metadata
// for a unary RPC.
func Trailer(md *metadata.MD) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*callOption); ok {
			o.opts = append(o.opts, grpc.Trailer(md))
		}
	}
}

func StreamDesc(streamDesc *grpc.StreamDesc) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*callOption); ok {
			o.streamDesc = streamDesc
		}
	}
}
