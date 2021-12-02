package rest

import (
	"net/http"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/base"
	"git.qutoutiao.net/gopher/qms/internal/invoker/rest"
)

// 远程调用时的路由方式
const (
	RouteDiscovery = "discovery" //指定采用服务发现
	RouteDirect    = "direct"    //指定采用SLB或ipport
	RouteSidecar   = "sidecar"   //指定采用sidecar
	RouteDefault   = ""          //自动判断。(如果url.Host含有"."，则判为SLB/ipport，否则，判为服务发现)
)

// rest调用时的可选参数
type CallOption = base.OptionFunc

// WithRoute用于明确指定路由类型, RouteDiscovery|RouteNormal|RouteSidecar|RouteDefault
func WithRoute(t string) CallOption {
	return rest.WithRoute(t)
}

// WithSidecar用于明确指定采用sidecar代理的方式访问远端
func WithSidecar() CallOption {
	return WithRoute(RouteSidecar)
}

// WithHeader用于添加header头
func WithHeader(h http.Header) CallOption {
	return rest.WithHeader(h)
}

// WithTimeout设置请求超时时间
func WithTimeout(timeout time.Duration) CallOption {
	return rest.WithTimeout(timeout)
}
