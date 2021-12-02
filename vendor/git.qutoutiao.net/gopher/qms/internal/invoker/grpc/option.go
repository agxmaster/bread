package grpc

import (
	"time"

	"git.qutoutiao.net/gopher/qms/balancer"
	"git.qutoutiao.net/gopher/qms/internal/base"
	cbalancer "git.qutoutiao.net/gopher/qms/internal/core/balancer"
)

type DialOption = base.OptionFunc
type CallOption = base.OptionFunc

type option struct {
	remoteService string // dial_option
	balancerName  string
	timeout       time.Duration // call_option
	routeType     string
}

func WithRemoteService(remoteService string) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			if remoteService != "" {
				o.remoteService = remoteService
			}
		}
	}
}

// 注册自定义LB
func WithBalancer(pb balancer.PickerBuilder) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			if pb != nil {
				cbalancer.RegisterBuilder(pb)
				o.balancerName = pb.Name()
			}
		}
	}
}

// 使用已注册的
func WithBalancerName(balancerName string) DialOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.balancerName = balancerName
		}
	}
}

// 配置请求超时时间
func WithTimeout(timeout time.Duration) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.timeout = timeout
		}
	}
}

// 配置请求超时时间
func WithRoute(t string) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.routeType = t
		}
	}
}
