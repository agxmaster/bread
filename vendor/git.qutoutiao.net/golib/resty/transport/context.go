package transport

import (
	"context"

	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/golib/resty/config"
)

const (
	ctxServiceKey ctxKey = iota
	ctxServiceNameKey
)

type ctxKey int

func ContextService(ctx context.Context) *config.ServiceConfig {
	value, ok := ctx.Value(ctxServiceKey).(*config.ServiceConfig)
	if !ok {
		return nil
	}

	return value
}

func ContextServiceName(ctx context.Context) string {
	value, ok := ctx.Value(ctxServiceNameKey).(string)
	if !ok {
		return ""
	}

	return value
}

func ContextWithService(ctx context.Context, service *config.ServiceConfig) context.Context {
	ctx = context.WithValue(ctx, ctxServiceKey, service)

	// for discovery balancer context
	ctx = balancer.ContextWithOption(ctx, balancer.CustomOption{
		DC:   service.DC,
		Tags: service.Tags,
	})

	return ctx
}

func ContextWithServiceName(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, ctxServiceNameKey, serviceName)
}
