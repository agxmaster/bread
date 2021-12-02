// Package client is an interface for any protocol's client
package client

import (
	"context"
	"errors"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
)

//ErrCanceled means Request is canceled by context management
var (
	ErrCanceled = errors.New("request cancelled")
	ErrTimeout  = errors.New("request timeout")
)

//TransportFailure is caused by client call failure
//for example:  resp, err = client.Do(req)
//if err is not nil then should wrap original error with TransportFailure
type TransportFailure struct {
	Message string
}

// Error return error message
func (e TransportFailure) Error() string {
	return e.Message
}

type CallOption = base.OptionFunc
type DialOption = base.OptionFunc

// ProtocolClient is the interface to communicate with one kind of ProtocolServer, it is used in transport handler
// rcp protocol client,http protocol client,or you can implement your own
type Client interface {
	// TODO use invocation.Response as rsp
	Call(ctx context.Context, addr string, inv *invocation.Invocation, rsp interface{}, opts ...CallOption) error
	String() string
	Close() error
	StatusCode(rsp interface{}, err error) int
	ReloadConfigs(Options) // 感觉没有必要
	GetOptions() Options   // 感觉没有必要
}
