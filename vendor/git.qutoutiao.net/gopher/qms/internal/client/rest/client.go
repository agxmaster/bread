package rest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.qutoutiao.net/golib/resty/transport"
	"git.qutoutiao.net/gopher/qms/internal/client/rest/httpcache"
	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/httputil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
)

const (
	// Name is a constant of type string
	Name = "http"
	// FailureTypePrefix is a constant of type string
	FailureTypePrefix = "http_"
	//DefaultTimeoutBySecond defines the default timeout for http connections
	DefaultTimeoutBySecond = 60 * time.Second
	//DefaultKeepAliveSecond defines the connection time
	DefaultKeepAliveSecond = 60 * time.Second
	//DefaultMaxConnsPerHost defines the maximum number of concurrent connections
	DefaultMaxConnsPerHost = 100
	//DefaultIdleConnTimeout is the maximum amount of time an idle connection will remain idle before closing  itself.
	DefaultIdleConnTimeout = 90 * time.Second
	//SchemaHTTP represents the http schema
	SchemaHTTP = "http"
	//SchemaHTTPS represents the https schema
	SchemaHTTPS = "https"
)

var (
	//ErrInvalidResp invalid input
	ErrInvalidResp = errors.New("rest consumer response arg is not *rest.Response type")
)

func init() {
	client.InstallPlugin(Name, New)
}

//Client is a struct
type Client struct {
	c     *http.Client
	opts  client.Options
	copts *CallOption
	//timeout time.Duration
	mu sync.Mutex // protects following
}

//NewRestClient is a function
func New(opts client.Options, dopts ...client.DialOption) (client.Client, error) {
	c := &Client{
		opts:  opts,
		copts: NewCallOption(opts.TLSConfig, opts.Failure),
	}

	dopts = append(dopts, WithPoolSize(opts.PoolSize), WithCredentials(opts.TLSConfig))

	// create http client
	if err := c.Dial(dopts...); err != nil {
		return nil, errors.WithStack(err)
	}

	return c, nil
}

func (c *Client) Dial(opts ...client.DialOption) (err error) {
	opt := newDialOption()
	for _, o := range opts {
		o(opt)
	}

	tp := &http.Transport{
		MaxIdleConns:        opt.poolSize,
		MaxIdleConnsPerHost: opt.poolSize,
		IdleConnTimeout:     DefaultIdleConnTimeout,
		DialContext: (&net.Dialer{
			KeepAlive: DefaultKeepAliveSecond,
			Timeout:   DefaultTimeoutBySecond,
		}).DialContext}
	if opt.tLSConfig != nil {
		tp.TLSClientConfig = opt.tLSConfig
	}

	var rt http.RoundTripper = transport.NewChainTransport(tp)
	if c.opts.RespCache > 0 {
		rt = &httpcache.Transport{
			Transport:           tp,
			Cache:               httpcache.NewCache(c.opts.RespCacheSize, c.opts.RespCache),
			MarkCachedResponses: true,
		}
	}

	c.c = &http.Client{
		Timeout:   c.opts.Timeout,
		Transport: rt,
	}

	return nil
}

// If a request fails, we generate an error.
func (c *Client) failure2Error(err error, resp *http.Response, addr string) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	if c.copts.failure == nil {
		return nil
	}

	// The Failure map defines whether or not a request fail.
	if c.copts.failure["http_"+strconv.Itoa(resp.StatusCode)] {
		return fmt.Errorf("http error status [%d], server addr [%s] (no response body)", resp.StatusCode, addr)
	}

	return nil
}

//Call is a method which uses client struct object
func (c *Client) Call(ctx context.Context, addr string, inv *invocation.Invocation, rsp interface{}, opts ...client.CallOption) (err error) {
	for _, o := range opts {
		o(c.copts)
	}

	reqSend, err := httputil.HTTPRequest(inv)
	if err != nil {
		return err
	}
	if addr != "" {
		reqSend.URL.Host = addr
		reqSend.Host = addr
	}

	resp, ok := rsp.(*http.Response)
	if !ok {
		return ErrInvalidResp
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//increase the max connection per host to prevent error "no free connection available" error while sending more requests.
	//TODO: check it
	//c.c.Transport.(*http.Transport).MaxIdleConnsPerHost = 512 * 20
	var temp *http.Response
	errChan := make(chan error, 1)
	go func() {
		// nolint: bodyclose
		temp, err = c.c.Do(reqSend)

		errChan <- errors.WithStack(err)
	}()

	select {
	case <-ctx.Done():
		err = errors.Wrapf(client.ErrCanceled, "%s: %s", reqSend.Method, reqSend.URL.String())
	case err = <-errChan:
		if err == nil {
			*resp = *temp
		}
	}

	return c.failure2Error(err, resp, addr)
}

func (c *Client) StatusCode(rsp interface{}, err error) (code int) {
	if err == nil {
		if resp, ok := rsp.(*http.Response); ok {
			return resp.StatusCode
		}
	}

	return http.StatusInternalServerError
}

func (c *Client) String() string {
	return "rest_client"
}

// Close is noop
func (c *Client) Close() error {
	return nil
}

// ReloadConfigs  reload configs for timeout and tls
func (c *Client) ReloadConfigs(opts client.Options) {
	c.opts = client.EqualOpts(c.opts, opts)
	c.copts = NewCallOption(c.opts.TLSConfig, c.opts.Failure)
}

// GetOptions method return opts
func (c *Client) GetOptions() client.Options {
	return c.opts
}
