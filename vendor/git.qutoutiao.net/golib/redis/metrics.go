package redis

import (
	"context"
	"time"

	"git.qutoutiao.net/golib/redis/metrics"
)

func (c *Client) metrics(ctx context.Context) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			stats := c.load().PoolStats()

			metrics.SetConnPool(c.config.Addr, stats)

			timer.Reset(time.Second)

		case <-ctx.Done():
			return
		}
	}
}
