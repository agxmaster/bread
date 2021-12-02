package registry

import (
	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
)

type Registry interface {
	Register(*Service, ...RegisterOption) error
	Deregister(protocol.Protocol, ...RegisterOption)
	Lookup(*ServiceKey) ([]*Service, error)
	Watch(*ServiceKey, Watcher)
	Close() error
}

// 注册器需要提供的option
type RegisterOption = base.OptionFunc

// call_back机制
type WatchFunc func(*ServiceKey, []*Service)

func (f WatchFunc) Handler(key *ServiceKey, services []*Service) {
	f(key, services)
}

type Watcher interface {
	Handler(*ServiceKey, []*Service)
}

//TODO: nop
