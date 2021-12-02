package weightedroundrobin

import (
	"sync/atomic"

	"github.com/golib/weighted"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

// wrrLoadBalance implements weighted round-robin alg.
type wrrLoadBalance struct {
	value atomic.Value
}

func (lb *wrrLoadBalance) next() (*registry.Service, error) {
	wrr, ok := lb.value.Load().(*weighted.RRW)
	if !ok {
		return nil, errors.ErrNotFound
	}

	service, ok := wrr.Next().(*registry.Service)
	if !ok {
		return nil, errors.ErrNotFound
	}

	return service, nil
}

func (lb *wrrLoadBalance) update(services []*registry.Service) {
	wrr := &weighted.RRW{}
	for _, service := range services {
		wrr.Add(service, service.ServiceWeight())
	}

	lb.value.Store(wrr)
}
