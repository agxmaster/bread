package consul

import (
	"git.qutoutiao.net/pedestal/discovery"
)

type Deregister interface {
	Deregister() error
	Name() string // TODO： 只是为了区分反注册的是哪个 后续改掉
}

type deregister struct {
	name            string // registry name
	ServiceRegister discovery.ServiceRegister
}

func newDeregister(name string, serviceRegister discovery.ServiceRegister) *deregister {
	return &deregister{
		name:            name,
		ServiceRegister: serviceRegister,
	}
}

func (d *deregister) Deregister() error {
	return d.ServiceRegister.Deregister()
}

func (d *deregister) Name() string {
	return d.name
}
