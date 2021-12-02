package balancer

import (
	"fmt"
	"strconv"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/registryutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qudiscovery"
	"git.qutoutiao.net/gopher/rebalancer"
	"git.qutoutiao.net/gopher/rebalancer/degrade"
	"github.com/hashicorp/consul/api"
)

type ReBalancerBuilder interface {
	Build() (ReBalancer, error)
}

type ReBalancer interface {
	Reload(servicename string, instances []*Service) error
	RecordMetrics(*Service, int)
	Next() (*Service, error)
}

type baseRbalancerBuilder struct {
	lbfunc rebalancer.NewLoadBalanceFunc
}

func newRebalancerBuilder(newLoadBalance rebalancer.NewLoadBalanceFunc) ReBalancerBuilder {
	return &baseRbalancerBuilder{
		lbfunc: newLoadBalance,
	}
}

func (bb *baseRbalancerBuilder) Build() (ReBalancer, error) {
	options := []rebalancer.RebalancerOption{rebalancer.RebalancerNew(bb.lbfunc)}
	if config.GetUpstream(constutil.Common).LoadBalance.WithoutBreaker {
		options = append(options, rebalancer.RebalancerNoBreaker())
	}
	rb, err := rebalancer.NewRebalancer(options...)
	if err != nil {
		return nil, err
	}

	return &wrapRebalance{
		rb: rb,
	}, nil
}

type wrapRebalance struct {
	rb *rebalancer.Rebalancer
}

func (wrb *wrapRebalance) RecordMetrics(instance *Service, code int) {
	wrb.rb.RecordMetrics(instance.ID, code)
}

func (wrb *wrapRebalance) Next() (*Service, error) {
	one, err := wrb.rb.Next()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	instance, ok := one.(*Service)
	if !ok {
		return nil, fmt.Errorf("convert [%v] to service failed", one)
	}
	return instance, nil
}

func (wrb *wrapRebalance) Reload(servicename string, instances []*Service) error {
	nodes := make([]*degrade.Node, 0, len(instances))
	for _, v := range instances {
		node := &degrade.Node{
			Address:   v.ID,
			Name:      v.Name,
			Host:      v.IP,
			Port:      strconv.FormatInt(int64(v.Port), 10),
			Weight:    int(v.Weight),
			IsOffline: registryutil.GetStatus(v.Meta) == registryutil.Offline,
			Node:      v,
		}
		if hc := registryutil.GetHealthCheck(v.Meta); hc != nil {
			interval, _ := time.ParseDuration(hc.Interval)
			timeout, _ := time.ParseDuration(hc.Timeout)
			node.HealthCheck = &qudiscovery.HealthCheck{
				Interval: api.ReadableDuration(interval),
				Timeout:  api.ReadableDuration(timeout),
				HTTP:     hc.HTTP,
				Header:   hc.Header,
				Method:   hc.Method,
				TCP:      hc.TCP,
			}
		}

		qlog.Debugf("Reload node: %+v", node)
		nodes = append(nodes, node)
	}
	return wrb.rb.Reload(servicename, nodes)
}
