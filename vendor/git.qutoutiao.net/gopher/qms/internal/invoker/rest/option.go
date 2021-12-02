package rest

import (
	"net/http"
	"time"

	"git.qutoutiao.net/gopher/qms/balancer"
	"git.qutoutiao.net/gopher/qms/internal/base"
	cbalancer "git.qutoutiao.net/gopher/qms/internal/core/balancer"
)

type DialOption = base.OptionFunc
type CallOption = base.OptionFunc

type option struct {
	balancerName string // dial_option
	routeType    string // call_option
	header       http.Header
	timeout      time.Duration
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

// WithRoute is a request option to specify route type. can be RouteDiscovery|RouteNormal|RouteSidecar|RouteDefault
func WithRoute(t string) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.routeType = t
		}
	}
}

func WithHeader(h http.Header) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.header = h
		}
	}
}

func WithTimeout(timeout time.Duration) CallOption {
	return func(opt interface{}) {
		if o, ok := opt.(*option); ok {
			o.timeout = timeout
		}
	}
}
