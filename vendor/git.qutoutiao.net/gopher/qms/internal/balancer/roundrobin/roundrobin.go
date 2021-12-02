package roundrobin

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qms/balancer"
	cbalancer "git.qutoutiao.net/gopher/qms/internal/core/balancer"
)

// Name is the name of round_robin balancer.
const Name = "RoundRobin"

func init() {
	rand.Seed(time.Now().UnixNano())
	cbalancer.RegisterBuilder(&rrPickerBuilder{})
}

type rrPickerBuilder struct{}

func (*rrPickerBuilder) Build(services []*balancer.Service) balancer.Picker {
	return &rrPicker{
		instances: services,
		next:      rand.Intn(len(services)),
	}
}

func (*rrPickerBuilder) Name() string {
	return Name
}

type rrPicker struct {
	instances []*balancer.Service
	next      int
	mu        sync.Mutex
}

func (p *rrPicker) Pick(ctx context.Context, opts *balancer.PickOptions) (*balancer.Service, error) {
	p.mu.Lock()
	sc := p.instances[p.next]
	p.next = (p.next + 1) % len(p.instances)
	p.mu.Unlock()
	return sc, nil
}
