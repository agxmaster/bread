package random

import (
	"context"
	"math/rand"
	"time"

	"git.qutoutiao.net/gopher/qms/balancer"
	cbalancer "git.qutoutiao.net/gopher/qms/internal/core/balancer"
)

// Name is the name of round_robin balancer.
const Name = "Random"

type rdPickBuilder struct{}

func (*rdPickBuilder) Build(services []*balancer.Service) balancer.Picker {
	return &rdPicker{
		instances: services,
	}
}

func (*rdPickBuilder) Name() string {
	return Name
}

type rdPicker struct {
	instances []*balancer.Service
}

func (p *rdPicker) Pick(ctx context.Context, opts *balancer.PickOptions) (*balancer.Service, error) {
	idx := rand.Intn(len(p.instances))
	return p.instances[idx], nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
	cbalancer.RegisterBuilder(&rdPickBuilder{})
}
