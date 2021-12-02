package discovery

import (
	"fmt"
	"strings"

	"git.qutoutiao.net/pedestal/discovery/registry"
)

type ServiceRegister interface {
	Deregister() error
}

type ServiceRegistrator struct {
	service      *registry.Service
	registrators []registry.Registrator
}

func (service *ServiceRegistrator) Deregister() (err error) {
	var errs []string
	for _, register := range service.registrators {
		err = register.Deregister(service.service)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%T.Deregister(%+v): %+v", register, service.service, err))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("%s", strings.Join(errs, ";"))
}
