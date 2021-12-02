package bwlist

import (
	"context"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/pedestal/discovery/balancer"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/logger/hclog"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

const (
	DefaultHeartbeatTimeout = time.Second
	DefaultCleanInterval    = int(time.Minute / DefaultHeartbeatTimeout) // 1m for recovery
)

func New(lb balancer.Balancer) balancer.Balancer {
	bwl := &bwlist{
		lb:         lb,
		log:        logger.NewWithHclog(hclog.NewWithSkipFrameCount(5)),
		black:      sync.Map{},
		heartbeats: sync.Map{},
	}
	go bwl.heartbeat()

	return bwl
}

// bwlist implements balancer.Balancer with black-white list policy.
type bwlist struct {
	lb         balancer.Balancer
	log        logger.Interface
	black      sync.Map // store all black services
	blacklen   uint32   // length of black list
	heartbeats sync.Map // store counter of heartbeat
}

func (bwl *bwlist) Next(ctx context.Context, name string) (service *registry.Service, err error) {
	if atomic.LoadUint32(&bwl.blacklen) == 0 {
		return bwl.lb.Next(ctx, name)
	}

	var counter uint32

	for {
		service, err = bwl.lb.Next(ctx, name)

		// fail-fast with lb error
		if err != nil {
			return
		}

		// the service is not a black one
		if _, ok := bwl.black.Load(service.ServiceID()); !ok {
			return
		}

		// exceeds length of black bwl
		if atomic.LoadUint32(&bwl.blacklen) < atomic.AddUint32(&counter, 1) {
			service = nil
			err = errors.ErrNotFound
			return
		}
	}
}

func (bwl *bwlist) Handle(key registry.ServiceKey, services []*registry.Service) {
	watcher, ok := bwl.lb.(registry.Watcher)
	if ok {
		watcher.Handle(key, services)
		return
	}

	bwl.log.Infof("bwlist.Handle(%+v, %+v): empty handler, ignored!", key, services)
	return
}

func (bwl *bwlist) Block(service *registry.Service) {
	if service == nil {
		return
	}

	bwl.black.Store(service.ServiceID(), service)
	atomic.AddUint32(&bwl.blacklen, 1)
}

func (bwl *bwlist) Unblock(service *registry.Service) {
	if service == nil {
		return
	}

	key := service.ServiceID()

	if _, ok := bwl.black.Load(key); !ok {
		return
	}

	bwl.unblock(key)
}

func (bwl *bwlist) heartbeat() {
	heartbeat := DefaultHeartbeatTimeout
	maxCounter := uint32(DefaultCleanInterval)

	timer := time.NewTimer(heartbeat)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			bwl.black.Range(func(key, value interface{}) bool {
				service, ok := value.(*registry.Service)
				if !ok {
					bwl.unblock(key)
				} else {
					conn, err := net.DialTimeout("tcp", service.Addr(), time.Second)
					if err != nil {
						bwl.log.Errorf("bwlist.Heartbeat(tcp, %s, 1s): %+v", service.Addr(), err)

						iface, loaded := bwl.heartbeats.LoadOrStore(key, uint32(1))
						if loaded {
							counter, ok := iface.(uint32)
							if ok {
								if atomic.AddUint32(&counter, 1) > maxCounter {
									bwl.unblock(key)
								} else {
									bwl.heartbeats.Store(key, counter)
								}
							} else {
								bwl.heartbeats.Store(key, uint32(1))
							}
						}
					} else {
						if err := conn.Close(); err != nil {
							bwl.log.Errorf("bwlist.Heartbeat(tcp, %s, 1s) close conn: %+v", service.Addr(), err)
						}

						bwl.unblock(key)
					}
				}

				return true
			})

			timer.Reset(heartbeat)
		}
	}
}

func (bwl *bwlist) unblock(key interface{}) {
	bwl.black.Delete(key)
	bwl.heartbeats.Delete(key)

	if atomic.AddUint32(&bwl.blacklen, ^uint32(0)) >= math.MaxUint32 {
		atomic.StoreUint32(&bwl.blacklen, 0)
	}
}
