package rest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	clientrest "git.qutoutiao.net/gopher/qms/internal/client/rest"
	restinvoker "git.qutoutiao.net/gopher/qms/internal/invoker/rest"
	"git.qutoutiao.net/gopher/qms/pkg/json"
)

func NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	return clientrest.NewRequest(method, urlStr, body)
}

// Do 采用context.Background() 不会传递Trace
func Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	return ContextDo(context.Background(), req, opts...)
}

// ContextDo类似于Do，只是多加了context参数用于传递trace
func ContextDo(ctx context.Context, req *http.Request, options ...CallOption) (*http.Response, error) {
	return restinvoker.NewInvoker().Invoke(ctx, req, options...)
}

// Get功能上类似于http.Get，框架内部附加了服务治理功能。（建议使用ContextGet）
func Get(url string, opts ...CallOption) (*http.Response, error) {
	return ContextGet(context.Background(), url, opts...)
}

// ContextGet等同于Get，只是多加了context参数用于传递trace
func ContextGet(ctx context.Context, url string, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return ContextDo(ctx, req, opts...)
}

// Post功能上类似于http.Post，框架内部附加了服务治理功能。（建议使用ContextPost）
func Post(url, contentType string, body io.Reader, opts ...CallOption) (*http.Response, error) {
	return ContextPost(context.Background(), url, contentType, body, opts...)
}

// ContextPost等同于Post，只是多加了context参数用于传递trace
func ContextPost(ctx context.Context, url, contentType string, body io.Reader, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return ContextDo(ctx, req, opts...)
}

// PostForm功能上类似于http.PostForm，框架内部附加了服务治理功能。（建议使用ContextPostForm）
func PostForm(url string, data url.Values, opts ...CallOption) (*http.Response, error) {
	return ContextPostForm(context.Background(), url, data, opts...)
}

// ContextPostForm等同于PostForm，只是多加了context参数用于传递trace
func ContextPostForm(ctx context.Context, url string, data url.Values, opts ...CallOption) (*http.Response, error) {
	return ContextPost(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), opts...)
}

// Head功能上类似于http.Head，框架内部附加了服务治理功能。（建议使用ContextHead）
func Head(url string, opts ...CallOption) (*http.Response, error) {
	return ContextHead(context.Background(), url, opts...)
}

// ContextHead等同于Head，只是多加了context参数用于传递trace
func ContextHead(ctx context.Context, url string, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return ContextDo(ctx, req, opts...)
}

// PostJson Post方法 使用application/json 自动序列化结构体
func PostJson(url string, input interface{}, opts ...CallOption) (*http.Response, error) {
	return ContextPostJson(context.Background(), url, input, opts...)
}

// ContextPostJson Post方法 使用application/json 自动序列化结构体
func ContextPostJson(ctx context.Context, url string, input interface{}, opts ...CallOption) (*http.Response, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return ContextPost(ctx, url, "application/json", bytes.NewBuffer(body), opts...)
}

type Invoker struct {
	i *restinvoker.Invoker
}

func NewInvoker(dopts ...DialOption) *Invoker {
	return &Invoker{
		i: restinvoker.NewInvoker(dopts...),
	}
}

// ContextDo is for requesting the API
func (i *Invoker) ContextDo(ctx context.Context, req *http.Request, options ...CallOption) (*http.Response, error) {
	return i.i.Invoke(ctx, req, options...)
}

func (i *Invoker) ContextGet(ctx context.Context, url string, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return i.ContextDo(ctx, req, opts...)
}

func (i *Invoker) ContextPost(ctx context.Context, url, contentType string, body io.Reader, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return i.ContextDo(ctx, req, opts...)
}

func (i *Invoker) ContextPostForm(ctx context.Context, url string, data url.Values, opts ...CallOption) (*http.Response, error) {
	return i.ContextPost(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), opts...)
}

func (i *Invoker) ContextHead(ctx context.Context, url string, opts ...CallOption) (*http.Response, error) {
	req, err := NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	return i.ContextDo(ctx, req, opts...)
}

// PostJson Post方法 使用application/json 自动序列化结构体
func (i *Invoker) PostJson(url string, input interface{}, opts ...CallOption) (*http.Response, error) {
	return ContextPostJson(context.Background(), url, input, opts...)
}

// ContextPostJson Post方法 使用application/json 自动序列化结构体
func (i *Invoker) ContextPostJson(ctx context.Context, url string, input interface{}, opts ...CallOption) (*http.Response, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return ContextPost(ctx, url, "application/json", bytes.NewBuffer(body), opts...)
}
