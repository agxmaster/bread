package invoker

import (
	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

type AbstractInvoker struct {
	dopts []client.DialOption
}

// new()
func NewInvoker(dopts ...client.DialOption) *AbstractInvoker {
	return &AbstractInvoker{
		dopts: dopts,
	}
}

// invoker
func (i *AbstractInvoker) Invoke(inv *invocation.Invocation, copts ...client.CallOption) (err error) {
	c, err := handler.GetChain(common.Consumer, common.DefaultChainName)
	if err != nil {
		qlog.Errorf("Handler chain init err [%s]", err.Error())
		return err
	}

	// add self service name into remote context, this value used in provider rate limiter
	inv.Ctx = common.WithContext(inv.Ctx, common.HeaderSourceName, runtime.ServiceName)

	// 根据用户制定的路由类型与服务名称，判断是否采用服务发现
	inv.NoDiscovery = rChecker.isNoDiscovery(inv.MicroServiceName, inv.GetRemoteService(), inv.RouteType)

	// 传递options
	inv.CallOptions = copts
	inv.DialOptions = i.dopts

	c.Next(inv, func(resp *invocation.Response) error {
		err = resp.Err
		return err
	})

	return
}

//func (i *AbstractInvoker) invoke(inv *invocation.Invocation, copts ...client.CallOption) invocation.Response {
//	c, err := client.GetClient(inv.Protocol, inv.MicroServiceName, inv.Endpoint, i.dopts...)
//	if err != nil {
//		qlog.Error("GetClient got Error: " + err.Error())
//		return invocation.Response{
//			Status: 500,
//			Err:    err,
//		}
//	}
//
//	//taking the time elapsed to check for latency aware strategy
//	timeBefore := time.Now()
//	err = c.Call(inv.Ctx, inv.Endpoint, inv, inv.Reply, copts...)
//	if err != nil {
//		if err != client.ErrCanceled {
//			qlog.Warnf("Call got Error, err [%s]", err.Error())
//		}
//
//		if inv.Strategy == loadbalancer.StrategySessionStickiness {
//			ProcessSpecialProtocol(inv)
//			ProcessSuccessiveFailure(inv)
//		}
//		return invocation.Response{
//			Result: inv.Reply,
//			Err:    err,
//			Status: c.StatusCode(inv.Reply, err),
//		}
//	}
//
//	if inv.Strategy == loadbalancer.StrategyLatency {
//		timeAfter := time.Since(timeBefore)
//		loadbalancer.SetLatency(timeAfter, inv.Endpoint, inv.MicroServiceName, inv.RouteTags, inv.Protocol)
//	}
//
//	if inv.Strategy == loadbalancer.StrategySessionStickiness {
//		ProcessSpecialProtocol(inv)
//	}
//
//	return invocation.Response{
//		Result: inv.Reply,
//		Status: c.StatusCode(inv.Reply, nil),
//	}
//}
//
////ProcessSpecialProtocol handles special logic for protocol
//func ProcessSpecialProtocol(inv *invocation.Invocation) {
//	switch inv.Protocol {
//	case common.ProtocolRest:
//		var reply *http.Response
//		if inv.Reply != nil && inv.Args != nil {
//			reply = inv.Reply.(*http.Response)
//			req := inv.Args.(*http.Request)
//			session.SaveSessionIDFromHTTP(inv.Endpoint, config.GetSessionTimeout(inv.SourceMicroService, inv.MicroServiceName), reply, req)
//		}
//	case common.ProtocolHighway:
//		inv.Ctx = session.SaveSessionIDFromContext(inv.Ctx, inv.Endpoint, config.GetSessionTimeout(inv.SourceMicroService, inv.MicroServiceName))
//	}
//}
//
////ProcessSuccessiveFailure handles special logic for protocol
//func ProcessSuccessiveFailure(inv *invocation.Invocation) {
//	var cookie string
//	var reply *http.Response
//
//	switch inv.Protocol {
//	case common.ProtocolRest:
//		if inv.Reply != nil && inv.Args != nil {
//			reply = inv.Reply.(*http.Response)
//		}
//		cookie = session.GetSessionCookie(nil, reply)
//		if cookie != "" {
//			loadbalancer.IncreaseSuccessiveFailureCount(cookie)
//			errCount := loadbalancer.GetSuccessiveFailureCount(cookie)
//			if errCount == config.StrategySuccessiveFailedTimes(inv.SourceServiceID, inv.MicroServiceName) {
//				session.DeletingKeySuccessiveFailure(reply)
//				loadbalancer.DeleteSuccessiveFailureCount(cookie)
//			}
//		}
//	default:
//		cookie = session.GetSessionCookie(inv.Ctx, nil)
//		if cookie != "" {
//			loadbalancer.IncreaseSuccessiveFailureCount(cookie)
//			errCount := loadbalancer.GetSuccessiveFailureCount(cookie)
//			if errCount == config.StrategySuccessiveFailedTimes(inv.SourceServiceID, inv.MicroServiceName) {
//				session.DeletingKeySuccessiveFailure(nil)
//				loadbalancer.DeleteSuccessiveFailureCount(cookie)
//			}
//		}
//	}
//}

// setCookieToCache   set go-chassisLB cookie to cache when use SessionStickiness strategy
//func setCookieToCache(inv invocation.Invocation, namespace string) {
//	if inv.Strategy != loadbalancer.StrategySessionStickiness {
//		return
//	}
//	cookie := session.GetSessionIDFromInv(inv, common.LBSessionID)
//	if cookie != "" {
//		cookies := strings.Split(cookie, "=")
//		if len(cookies) > 1 {
//			session.AddSessionStickinessToCache(cookies[1], namespace)
//		}
//	}
//}

// getNamespaceFromMetadata get namespace from opts.Metadata
//func getNamespaceFromMetadata(metadata map[string]interface{}) string {
//	if namespaceTemp, ok := metadata[common.SessionNameSpaceKey]; ok {
//		if v, ok := namespaceTemp.(string); ok {
//			return v
//		}
//	}
//	return common.SessionNameSpaceDefaultValue
//}
