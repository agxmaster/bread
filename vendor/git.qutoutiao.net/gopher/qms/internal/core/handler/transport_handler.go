package handler

import (
	"net/http"

	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/loadbalancer"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/internal/session"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

// TransportHandler transport handler
type TransportHandler struct{}

func newTransportHandler() Handler {
	return &TransportHandler{}
}

// Name returns transport string
func (th *TransportHandler) Name() string {
	return "transport"
}

// Handle is to handle transport related things
func (th *TransportHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	c, err := client.GetClient(i.Protocol.String(), i.MicroServiceName, i.Endpoint, i.DialOptions...)
	if err != nil {
		errNotNil(err, cb)
		return
	}

	r := &invocation.Response{}

	//taking the time elapsed to check for latency aware strategy
	//timeBefore := time.Now()
	err = c.Call(i.Ctx, i.Endpoint, i, i.Reply, i.CallOptions...)
	r.Status = c.StatusCode(i.Reply, err)
	if err != nil {
		r.Err = err
		qlog.WithError(err).Errorf("endpoint: %s", i.Endpoint)
		if i.Strategy == loadbalancer.StrategySessionStickiness {
			ProcessSpecialProtocol(i)
			ProcessSuccessiveFailure(i)
		}
		cb(r)
		return
	}

	//if i.Strategy == loadbalancer.StrategyLatency {
	//	timeAfter := time.Since(timeBefore)
	//	loadbalancer.SetLatency(timeAfter, i.Endpoint, i.MicroServiceName, i.RouteTags, i.Protocol)
	//}

	if i.Strategy == loadbalancer.StrategySessionStickiness {
		ProcessSpecialProtocol(i)
	}

	r.Result = i.Reply
	cb(r)
}

//ProcessSpecialProtocol handles special logic for protocol
func ProcessSpecialProtocol(inv *invocation.Invocation) {
	switch inv.Protocol {
	case protocol.ProtocHTTP:
		var reply *http.Response
		if inv.Reply != nil && inv.Args != nil {
			reply = inv.Reply.(*http.Response)
			req := inv.Args.(*http.Request)
			session.SaveSessionIDFromHTTP(inv.Endpoint, config.GetSessionTimeout(inv.SourceMicroService, inv.MicroServiceName), reply, req)
		}
	}
}

//ProcessSuccessiveFailure handles special logic for protocol
func ProcessSuccessiveFailure(i *invocation.Invocation) {
	var cookie string
	var reply *http.Response

	switch i.Protocol {
	case protocol.ProtocHTTP:
		if i.Reply != nil && i.Args != nil {
			reply = i.Reply.(*http.Response)
		}
		cookie = session.GetSessionCookie(nil, reply)
		if cookie != "" {
			loadbalancer.IncreaseSuccessiveFailureCount(cookie)
			errCount := loadbalancer.GetSuccessiveFailureCount(cookie)
			if errCount == config.StrategySuccessiveFailedTimes(i.SourceServiceID, i.MicroServiceName) {
				session.DeletingKeySuccessiveFailure(reply)
				loadbalancer.DeleteSuccessiveFailureCount(cookie)
			}
		}
	default:
		cookie = session.GetSessionCookie(i.Ctx, nil)
		if cookie != "" {
			loadbalancer.IncreaseSuccessiveFailureCount(cookie)
			errCount := loadbalancer.GetSuccessiveFailureCount(cookie)
			if errCount == config.StrategySuccessiveFailedTimes(i.SourceServiceID, i.MicroServiceName) {
				session.DeletingKeySuccessiveFailure(nil)
				loadbalancer.DeleteSuccessiveFailureCount(cookie)
			}
		}
	}
}

func errNotNil(err error, cb invocation.ResponseCallBack) {
	r := &invocation.Response{
		Err: err,
	}
	qlog.Errorf("GetClient got Error: " + err.Error())
	cb(r)
	return
}
