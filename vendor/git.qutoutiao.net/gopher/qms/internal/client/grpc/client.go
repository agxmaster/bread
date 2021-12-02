package grpc

import (
	"context"
	"errors"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func init() {
	client.InstallPlugin("grpc", New)
}

const errPrefix = "grpc client: "

//Client is grpc client holder
type Client struct {
	c           *grpc.ClientConn
	opts        client.Options // 仅提供给ReloadConfigs接口使用
	service     string
	callTimeout time.Duration
}

// New create new grpc client
func New(opts client.Options, dopts ...client.DialOption) (client.Client, error) {
	c := &Client{
		opts:        opts,
		service:     opts.Service,
		callTimeout: opts.Timeout,
	}

	if opts.TLSConfig == nil {
		dopts = append(dopts, WithInsecure())
	} else {
		dopts = append(dopts, WithTransportCredentials(credentials.NewTLS(opts.TLSConfig)))
	}

	if err := c.Dial(opts.Endpoint, dopts...); err != nil {
		err = errors.New(errPrefix + err.Error())
		return nil, err
	}

	return c, nil
}

func (c *Client) Dial(target string, opts ...DialOption) (err error) {
	opt := &dialOption{}
	for _, o := range opts {
		o(opt)
	}

	// make interceptor chain
	opt.opts = append(opt.opts, grpc.WithChainUnaryInterceptor(opt.unaryInts...))
	opt.opts = append(opt.opts, grpc.WithChainStreamInterceptor(opt.streamInts...))

	ctx := context.Background()
	if opt.dialTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.dialTimeout)
		defer cancel()
	}
	c.c, err = grpc.DialContext(ctx, target, opt.opts...)
	if err != nil {
		return err
	}
	return
}

//TransformContext will deliver header in chassis context key to grpc context key
func TransformContext(ctx context.Context) context.Context {
	header := common.FromContext(ctx)
	kvs := make([]string, 0, header.Len())
	for k, vv := range header {
		for _, v := range vv {
			kvs = append(kvs, k, v)
		}
	}
	return metadata.AppendToOutgoingContext(ctx, kvs...) // grpc推荐的方式
}

//Call remote server
func (c *Client) Call(ctx context.Context, addr string, inv *invocation.Invocation, rsp interface{}, opts ...client.CallOption) (err error) {
	opt := &callOption{}
	for _, o := range opts {
		o(opt)
	}

	ctx = TransformContext(ctx)
	if opt.streamDesc == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.callTimeout)
		err = c.c.Invoke(ctx, inv.OperationID, inv.Args, rsp, opt.opts...)
		cancel()
	} else {
		var stream grpc.ClientStream
		stream, err = c.c.NewStream(ctx, opt.streamDesc, inv.OperationID, opt.opts...)
		if stream != nil {
			if resp, ok := rsp.(*grpc.ClientStream); ok {
				*resp = stream
			}
		}
	}
	return
}

func (c *Client) StatusCode(rsp interface{}, err error) int {
	return runtime.HTTPStatusFromCode(status.Code(err))
}

//String return name
func (c *Client) String() string {
	return "grpc_client"
}

// Close close conn
func (c *Client) Close() error {
	return c.c.Close()
}

// ReloadConfigs reload configs for timeout and tls
func (c *Client) ReloadConfigs(opts client.Options) {
	newOpts := client.EqualOpts(c.opts, opts)
	if newOpts.TLSConfig != c.opts.TLSConfig {
		conn, err := New(opts)
		if err == nil && conn != nil {
			if c.c != nil {
				c.c.Close()
			}

			c.c = conn.(*Client).c
		}
	}
}

//GetOptions method return opts
func (c *Client) GetOptions() client.Options {
	return c.opts
}
