package handler

import (
	"fmt"

	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	stringutil "git.qutoutiao.net/gopher/qms/internal/pkg/string"
)

// Handler interface for handlers
type Handler interface {
	// handle invocation transportation,and tr response
	Handle(*Chain, *invocation.Invocation, invocation.ResponseCallBack)
	Name() string
}

// RegisterHandler Let developer custom handler
func RegisterHandler(name string, f func() Handler) error {
	if stringutil.StringInSlice(name, buildInHandlers) {
		return errViolateBuildIn
	}
	_, ok := handlerStore[name]
	if ok {
		return ErrDuplicatedHandler
	}
	handlerStore[name] = f
	return nil
}

// CreateHandler create a new handler by name your registered
func CreateHandler(name string) (Handler, error) {
	f := handlerStore[name]
	if f == nil {
		return nil, fmt.Errorf("don't have handler [%s]", name)
	}
	return f(), nil
}

func writeErr(err error, cb invocation.ResponseCallBack) {
	r := &invocation.Response{
		Err: err,
	}

	cb(r)
}
