package handler

import (
	"fmt"
	"net/http"

	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/qpslimiter"
	pkgerr "git.qutoutiao.net/gopher/qms/pkg/errors"
)

// ProviderRateLimiterHandler provider rate limiter handler
type ProviderRateLimiterHandler struct{}

// Handle is to handle provider rateLimiter things
func (rl *ProviderRateLimiterHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	rlc := control.DefaultPanel.GetRateLimiting(i, common.Provider)
	if !rlc.Enabled {
		chain.Next(i, cb)

		return
	}

	limited := false
	if rlc.Rate <= 0 {
		limited = true
	} else {
		allowed := qpslimiter.GetQPSTrafficLimiter().ProcessQPSTokenReq(rlc.Key, rlc.Rate)
		if !allowed {
			limited = true
		}
	}

	if limited {
		// ignore /ping, /metrics
		if i.URLPathFormat == common.DefaultHealthzPath || i.URLPathFormat == common.DefaultMetricsPath {
			limited = false
		}
	}

	if limited {
		switch i.Reply.(type) {
		case *http.Response:
			resp := i.Reply.(*http.Response)
			resp.StatusCode = http.StatusTooManyRequests
		}

		r := &invocation.Response{}
		r.Status = http.StatusTooManyRequests
		r.Err = pkgerr.WithMessage(pkgerr.ErrRateLimit, fmt.Sprintf("ratelimit: %s|%v", rlc.Key, rlc.Rate))
		cb(r)
		return
	}

	//call next chain
	chain.Next(i, cb)
}

func newProviderRateLimiterHandler() Handler {
	return &ProviderRateLimiterHandler{}
}

// Name returns the name providerratelimiter
func (rl *ProviderRateLimiterHandler) Name() string {
	return "providerratelimiter"
}
