package rest

import (
	"git.qutoutiao.net/gopher/qms/balancer"
	"git.qutoutiao.net/gopher/qms/internal/base"
	invokerrest "git.qutoutiao.net/gopher/qms/internal/invoker/rest"
)

type DialOption = base.OptionFunc

// WithBalancerName用于指定负载均衡类型
func WithBalancerName(balancerName string) DialOption {
	return invokerrest.WithBalancerName(balancerName)
}

// WithBalancer用于指定自定义的负载均衡策略
func WithBalancer(pb balancer.PickerBuilder) DialOption {
	return invokerrest.WithBalancer(pb)
}
