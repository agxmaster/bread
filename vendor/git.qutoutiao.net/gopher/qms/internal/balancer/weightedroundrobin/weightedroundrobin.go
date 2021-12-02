package weightedroundrobin

import (
	"sync/atomic"

	"git.qutoutiao.net/gopher/qms/internal/core/balancer"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"github.com/golib/weighted"
)

// Name is the name of round_robin balancer.
const Name = "WeightedRoundRobin"

func init() {
	balancer.RegisterLoadbalancerBuilder(&wrrPickBuilder{})
}

type wrrPickBuilder struct{}

func (*wrrPickBuilder) Build() balancer.Loadbalancer {
	var (
		picker = &wrrPicker{}
		rrw    = &weighted.RRW{}
	)

	picker.value.Store(rrw)
	return picker
}

func (*wrrPickBuilder) Name() string {
	return Name
}

type wrrPicker struct {
	value atomic.Value
}

func (wrr *wrrPicker) Next() (interface{}, error) {
	rrw, ok := wrr.value.Load().(*weighted.RRW)
	if !ok {
		return nil, balancer.ErrNotFound
	}

	service, ok := rrw.Next().(*balancer.Service)
	if !ok {
		return nil, balancer.ErrNotFound
	}

	return service, nil
}

// 只创建 不更新
func (wrr *wrrPicker) Upsert(id string, weight int, service interface{}) error {
	rrw, ok := wrr.value.Load().(*weighted.RRW)
	if !ok {
		return balancer.ErrNotFound
	}
	rrw.Add(service, weight)
	qlog.Debugf("wrr add id[%s], weight[%d]", id, weight)
	return nil
}

// shuffleService防止每次都从固定的位置开始
//func shuffleService(services []*balancer.Service) {
//	if len(services) < 2 {
//		return
//	}
//
//	rand.Seed(time.Now().UnixNano())
//
//	rand.Shuffle(len(services), func(i, j int) {
//		services[i], services[j] = services[j], services[i]
//	})
//}

//func (wrr *wrrPicker) Pick(ctx context.Context, opts *balancer.PickOptions) (*balancer.Service, error) {
//	rrw, ok := wrr.value.Load().(*weighted.RRW)
//	if !ok {
//		return nil, balancer.ErrNotFound
//	}
//
//	service, ok := rrw.Next().(*balancer.Service)
//	if !ok {
//		return nil, balancer.ErrNotFound
//	}
//
//	return service, nil
//}
