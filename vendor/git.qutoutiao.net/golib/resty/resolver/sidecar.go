package resolver

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/discovery/eds"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
	"golang.org/x/sync/singleflight"
)

const (
	DefaultSidecarServiceID = "sidecar"
	DefaultSidecarAddr      = "http://127.0.0.1:8102"
)

const (
	SidecarHeaderDatacenterKey = "X-Qtt-Meshdc"
	SidecarHeaderServiceKey    = "X-Qtt-Meshservice"
	SidecarHeaderTagsKey       = "X-Qtt-Meshtags"
	SidecarHeaderErrorKey      = "X-Qtt-Mesherror"
)

type sidecarRecord struct {
	sidecarAddr   string
	sidecarPort   int
	sidecarValue  *atomic.Value
	resolverValue *atomic.Value
	single        *singleflight.Group
	store         sync.Map
}

func NewSidecarResolver(addr string) Interface {
	return NewSidecarResolverWithEDS(addr, "")
}

func NewSidecarResolverWithEDS(sidecarAddr, edsAddr string) Interface {
	if len(sidecarAddr) == 0 {
		sidecarAddr = DefaultSidecarAddr
	}

	urlobj, err := url.Parse(sidecarAddr)
	if err == nil && len(urlobj.Host) > 0 {
		sidecarAddr = urlobj.Host
	}

	sidecar := &sidecarRecord{
		sidecarValue:  new(atomic.Value),
		resolverValue: new(atomic.Value),
		single:        new(singleflight.Group),
	}

	// apply sidecar value
	tcpAddr, tcpErr := net.ResolveTCPAddr("tcp", sidecarAddr)
	if tcpErr != nil {
		sidecar.sidecarValue.Store(false)
	} else {
		sidecar.sidecarValue.Store(true)

		sidecar.sidecarAddr = tcpAddr.IP.String()
		sidecar.sidecarPort = tcpAddr.Port
	}

	// apply eds value
	if len(edsAddr) > 0 {
		edsAdapter, edsErr := eds.NewWithInterval(edsAddr, 10*time.Second)
		if edsErr == nil {
			sidecar.resolverValue.Store(NewRegistryResolver(discovery.NewRegistry(discovery.WithDiscoveries(edsAdapter))))
		} else {
			sidecar.resolverValue.Store(NewRegistryResolver(nil, edsErr))
		}
	}

	go sidecar.heartbeat()

	return sidecar
}

func (record *sidecarRecord) Resolve(ctx context.Context, name string) (service *registry.Service, err error) {
	err = ctx.Err()
	if err != nil {
		return
	}

	value, ok := record.sidecarValue.Load().(bool)
	if ok && value {
		service = &registry.Service{
			ID:     DefaultSidecarServiceID,
			Name:   name,
			IP:     record.sidecarAddr,
			Port:   record.sidecarPort,
			Weight: 100,
		}

		return
	}

	// set not found by default
	err = errors.ErrNotFound

	// try fall-back resolver
	resolver, ok := record.resolverValue.Load().(Interface)
	if !ok || resolver == nil {
		return
	}

	return resolver.Resolve(ctx, name)
}

func (record *sidecarRecord) Block(ctx context.Context, name string, service *registry.Service) {
	switch service.ID {
	case DefaultSidecarServiceID: // for sidecar connect
		record.sidecarValue.Store(false)

	default:
		resolver, ok := record.resolverValue.Load().(Interface)
		if !ok || resolver == nil {
			return
		}

		resolver.Block(ctx, name, service)
	}
}

func (record *sidecarRecord) Close() {
	resolver, ok := record.resolverValue.Load().(Interface)
	if !ok || resolver == nil {
		return
	}

	resolver.Close()
}

func (record *sidecarRecord) heartbeat() {
	heartbeat := 3 * time.Second

	timer := time.NewTimer(heartbeat)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			value, ok := record.sidecarValue.Load().(bool)
			if ok && value {
				timer.Reset(heartbeat)
				continue
			}

			conn, err := net.DialTimeout("tcp", record.sidecarAddr+":"+strconv.Itoa(record.sidecarPort), time.Second)
			if err == nil {
				conn.Close()

				record.sidecarValue.Store(true)
			}

			timer.Reset(heartbeat)
		}
	}
}
