package handler

import (
	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/balancer"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

// LBHandler loadbalancer handler struct
type LBHandler struct{}

// TODO: 要针对服务 而不是 common
func newLBHandler() Handler {
	return &LBHandler{}
}

// Name returns loadbalancer string
func (lb *LBHandler) Name() string {
	return "loadbalancer"
}

// Handle to handle the request
func (lb *LBHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	// do not using discovery, so skipping LB
	if i.NoDiscovery {
		i.Endpoint = i.GetRemoteService()

		if sidecar := i.GetUpstream().Sidecar; i.RouteType == common.RouteSidecar && sidecar.Enabled {
			destName := i.GetMeshService()
			i.Ctx = common.WithContext(i.Ctx, common.HeaderXSidecar, destName)
			i.Endpoint = sidecar.Address
			qlog.Debugf("request router=%s, service=%s", common.RouteSidecar, common.HeaderXSidecar, destName)
		} else {
			qlog.Debugf("request router=default, endpoint=%s", i.Endpoint)
		}

		chain.Next(i, cb)
		return
	}

	lb.handleWithLB(chain, i, cb)
}

func (lb *LBHandler) handleWithLB(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	// should run with lb
	lbConfig := control.DefaultPanel.GetLoadBalancing(i)
	if !lbConfig.Enabled {
		chain.Next(i, cb)
		return
	}

	rb, err := lb.getReBalancer(i, lbConfig)
	if err != nil {
		writeErr(err, cb)
		return
	}

	one, err := rb.Next()
	if err != nil {
		writeErr(err, cb)
		return
	}

	i.Endpoint = one.Endpoint
	qlog.Debugf("request router=%T, endpoint=%s", rb, i.Endpoint)

	chain.Next(i, func(r *invocation.Response) (err error) {
		rb.RecordMetrics(one, r.Status)
		return cb(r)
	})
}

func (lb *LBHandler) getReBalancer(i *invocation.Invocation, lbConfig control.LoadBalancingConfig) (balancer.ReBalancer, error) {
	var err error
	if i.Strategy == "" {
		i.Strategy = lbConfig.Strategy
	}
	if len(i.Filters) == 0 {
		i.Filters = lbConfig.Filters
	}

	b, err := balancer.GetBalancer(i.Strategy)
	if err != nil {
		return nil, err
	}

	return b.ReBalancer(i.Ctx, balancer.NewOptions(i))
}
