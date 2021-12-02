package archaius

import (
	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/loadbalancer"
	"git.qutoutiao.net/gopher/qms/internal/pkg/backoff"
	"github.com/patrickmn/go-cache"
)

//save configs
var (
	//key is service name
	LBConfigCache = cache.New(0, 0)
	//key is [Provider|Consumer]:service_name or [Consumer|Provider]
	CBConfigCache     = cache.New(0, 0)
	RLConfigCache     = cache.New(0, 0)
	EgressConfigCache = cache.New(0, 0)
	FIConfigCache     = cache.New(0, 0)
)

//Default values
var (
	DefaultLB = control.LoadBalancingConfig{
		Strategy:    loadbalancer.StrategyWeightedRoundRobin,
		BackOffKind: backoff.DefaultBackOffKind,
	}
)
