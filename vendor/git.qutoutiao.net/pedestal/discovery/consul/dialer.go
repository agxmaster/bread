package consul

import (
	"context"
	"net"
)

type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func WrapDialContext(host string, origin DialFunc) DialFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 先尝试返回一个特定的地址
		conn, err := origin(ctx, network, host)
		if err == nil {
			return conn, nil
		}

		return origin(ctx, network, addr)
	}
}
