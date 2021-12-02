package grpc

import (
	"time"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DialOption = base.OptionFunc
type dialOption struct {
	dialTimeout time.Duration
	unaryInts   []grpc.UnaryClientInterceptor
	streamInts  []grpc.StreamClientInterceptor
	opts        []grpc.DialOption // 起到过滤和拦截的作用，qms暴露的才能传递
}

// WithBlock returns a DialOption which makes caller of Dial blocks until the
// underlying connection is up. Without this, Dial returns immediately and
// connecting the server happens in background.
func WithBlock() DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.opts = append(o.opts, grpc.WithBlock())
		}
	}
}

// WithInsecure returns a DialOption which disables transport security for this
// ClientConn. Note that transport security is required unless WithInsecure is
// set.
func WithInsecure() DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.opts = append(o.opts, grpc.WithInsecure())
		}
	}
}

func WithDialTimeout(d time.Duration) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.dialTimeout = d
		}
	}
}

// WithTransportCredentials returns a DialOption which configures a connection
// level security credentials (e.g., TLS/SSL). This should not be used together
// with WithCredentialsBundle.
func WithTransportCredentials(creds credentials.TransportCredentials) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.opts = append(o.opts, grpc.WithTransportCredentials(creds))
		}
	}
}

func WithUnaryClientChain(interceptors ...grpc.UnaryClientInterceptor) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.unaryInts = append(o.unaryInts, interceptors...)
		}
	}
}

func WithStreamClientChain(interceptors ...grpc.StreamClientInterceptor) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*dialOption); ok {
			o.streamInts = append(o.streamInts, interceptors...)
		}
	}
}
