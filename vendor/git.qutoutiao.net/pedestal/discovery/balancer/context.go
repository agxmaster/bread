package balancer

import (
	"context"

	"git.qutoutiao.net/pedestal/discovery/errors"
)

var (
	loadBalanceOption optionKey
)

type optionKey struct{}

func ContextWithOption(ctx context.Context, opt CustomOption) context.Context {
	return context.WithValue(ctx, loadBalanceOption, opt)
}

func OptionFromContext(ctx context.Context) (CustomOption, error) {
	v := ctx.Value(loadBalanceOption)
	if v == nil {
		return CustomOption{}, errors.ErrNotFound
	}

	if opt, ok := v.(CustomOption); ok {
		return opt, nil
	}

	return CustomOption{}, errors.ErrNotFound
}
