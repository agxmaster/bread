package rolling

import (
	"sync"
	"time"

	"git.qutoutiao.net/pedestal/discovery/logger"

	"golang.org/x/sync/singleflight"
)

type Singleflight struct {
	ctype   string
	timeout time.Duration
	latency time.Duration
	logger  logger.Interface

	single *singleflight.Group
	store  sync.Map
}

func NewSingleflight(ctype string, timeout, latency time.Duration, logger logger.Interface) *Singleflight {
	return &Singleflight{
		ctype:   ctype,
		timeout: timeout,
		latency: latency,
		logger:  logger,
		single:  new(singleflight.Group),
		store:   sync.Map{},
	}
}

func (c *Singleflight) Delay(key string) {
	c.flight(key).Delay(c.ctype, key)
}

func (c *Singleflight) DelayWithShift(key string) {
	c.flight(key).DelayWithShift(c.ctype, key)
}

func (c *Singleflight) flight(key string) *Backoff {
	iface, _, _ := c.single.Do(key, func() (v interface{}, err error) {
		v, ok := c.store.Load(key)
		if !ok {
			// create a new backoff
			v = NewBackoffWithLogger(c.timeout, c.latency, c.logger)

			c.store.Store(key, v)
		}

		return
	})

	backoff, ok := iface.(*Backoff)
	if !ok {
		backoff = NewBackoffWithLogger(c.timeout, c.latency, c.logger)

		c.store.Store(key, backoff)
	}

	return backoff
}
