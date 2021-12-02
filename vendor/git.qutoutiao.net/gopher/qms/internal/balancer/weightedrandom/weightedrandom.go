package weightedrandom

import (
	"context"
	"sync/atomic"

	"git.qutoutiao.net/gopher/qms/balancer"
	cbalancer "git.qutoutiao.net/gopher/qms/internal/core/balancer"
	"github.com/golib/weighted"
)

// Name is the name of random balancer.
const Name = "WeightedRandom"

func init() {
	cbalancer.RegisterBuilder(&wrdPickBuilder{})
}

type wrdPickBuilder struct{}

func (*wrdPickBuilder) Build(services []*balancer.Service) balancer.Picker {
	var (
		picker = &wrdPicker{}
		wrand  = weighted.NewRandW()
	)
	for _, service := range services {
		wrand.Add(service, int(service.Weight))
	}
	picker.value.Store(wrand)
	return picker
}

func (*wrdPickBuilder) Name() string {
	return Name
}

type wrdPicker struct {
	value atomic.Value
}

func (wrd *wrdPicker) Pick(ctx context.Context, opts *balancer.PickOptions) (*balancer.Service, error) {
	wrand, ok := wrd.value.Load().(*weighted.RandW)
	if !ok {
		return nil, cbalancer.ErrNotFound
	}

	service, ok := wrand.Next().(*balancer.Service)
	if !ok {
		return nil, cbalancer.ErrNotFound
	}

	return service, nil
}
