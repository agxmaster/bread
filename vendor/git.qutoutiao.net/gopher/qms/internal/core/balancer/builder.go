package balancer

import (
	"sync"

	qmsbalancer "git.qutoutiao.net/gopher/qms/balancer"
	"git.qutoutiao.net/gopher/rebalancer"
)

// Builder creates a balancer.
type Builder interface {
	// Build a balancer
	Build(resolver Resolver) Balancer
}

// builder is a map from name to balancer builder.
var builders sync.Map // map[name]Builder

func RegisterBuilder(pb qmsbalancer.PickerBuilder) {
	//builders.Store(pb.Name(), newBalancerBuilder(pb))
}

func RegisterLoadbalancerBuilder(pb LoadbalancerBuilder) {
	builders.Store(pb.Name(), newBalancerBuilder(
		newRebalancerBuilder(func() rebalancer.LoadBalance {
			return pb.Build()
		})))
}

func getBuilder(name string) Builder {
	if b, ok := builders.Load(name); ok {
		return b.(Builder)
	}
	return nil
}
