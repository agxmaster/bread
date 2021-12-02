package balancer

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"git.qutoutiao.net/gopher/qms/internal/core/registry"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"golang.org/x/sync/singleflight"
)

var (
	ErrArgument        = errors.New("error argument")
	ErrNotFound        = errors.New("not found")
	ErrLookup          = errors.New("lookup service from registry failed")
	ErrOnChanged       = errors.New("services changed failed")
	ErrBuildRebalancer = errors.New("build rebalancer failed")
)

// NewBalancerBuilderWithConfig returns a base balancer builder configured by the provided config.
func newBalancerBuilder(rbuilder ReBalancerBuilder) Builder {
	return &baseBuilder{
		rbuilder: rbuilder,
	}
}

type baseBuilder struct {
	rbuilder ReBalancerBuilder
}

func (bb *baseBuilder) Build(resolver Resolver) Balancer {
	return &baseBalancer{
		resolver:     resolver,
		rbuilder:     bb.rbuilder,
		singleflight: new(singleflight.Group),
	}
}

type baseBalancer struct {
	store        sync.Map // map[ServiceKey]ReBalancer
	resolver     Resolver
	rbuilder     ReBalancerBuilder
	singleflight *singleflight.Group
}

func (b *baseBalancer) OnChanged(key *registry.ServiceKey, instances []*Service) error {
	value, ok := b.store.Load(key.String())
	if !ok {
		return errors.Newf("can't load rebalancer by key(%s)", key.String())
	}

	rb, ok := value.(ReBalancer)
	if !ok {
		return errors.Newf("value(%v) can't convert rebalancer", value)
	}
	if err := rb.Reload(key.Name(), instances); err != nil {
		return errors.WithStack(err)
	}

	builder := strings.Builder{}
	for _, v := range instances {
		builder.WriteString(v.Endpoint)
		builder.WriteString(":")
		builder.WriteString(strconv.FormatInt(int64(v.Weight), 10))
		builder.WriteString(" ")
	}
	qlog.Infof("key[%s]获取实例[%s]", key.String(), builder.String())

	return nil
}

func (b *baseBalancer) ReBalancer(ctx context.Context, opts *Options) (ReBalancer, error) {
	key := registry.NewServiceKey(opts.RemoteService, opts.Datacenter, opts.Tags, opts.Protocol, opts.Env)
	keyStr := key.String()
	rbalancer, ok := b.store.Load(keyStr)
	if ok {
		return rbalancer.(ReBalancer), nil
	}

	picker, err, _ := b.singleflight.Do(keyStr, func() (interface{}, error) {
		services, err := b.resolver.Lookup(key)
		if err != nil {
			return nil, errors.Wrapf(ErrLookup, "err: %s", err.Error())
		}
		if len(services) <= 0 {
			return nil, ErrNotFound
		}
		qlog.Tracef("resolver find [%d] services by [%s]", len(services), key.String())

		// 创建rb
		rbalancer, err := b.rbuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(ErrBuildRebalancer, "err: %s", err.Error())
		}
		b.store.Store(keyStr, rbalancer)
		// 初始化rb
		if err = b.OnChanged(key, services); err != nil {
			qlog.WithError(err).Error("OnChanged error.")
		}
		// watch
		b.resolver.WithWatchFunc(key, WatchFunc(func(yek *registry.ServiceKey, instances []*Service) {
			b.OnChanged(key, instances) //注:因为consul发现时"_""-"替换的原因，name与serviceName的值可能不完全一致
		}))
		return rbalancer, nil
	})
	if err != nil {
		return nil, err
	}

	if rb, ok := picker.(ReBalancer); ok {
		return rb, nil
	}

	return nil, ErrArgument
}

// NewErrPicker returns a picker that always returns err on Pick().
//func NewErrPicker(err error) qmsbalancer.Picker {
//	return &errPicker{err: err}
//}
//
//type errPicker struct {
//	err error // Pick() always returns this err.
//}
//
//func (p *errPicker) Pick(ctx context.Context, opts *qmsbalancer.PickOptions) (*qmsbalancer.Service, error) {
//	return nil, p.err
//}

//func (b *baseBalancer) Pick(ctx context.Context, opts *qmsbalancer.PickOptions) (*qmsbalancer.Service, error) {
//	key := registry.NewServiceKey(opts.RemoteService, registry.WithDC(opts.Datacenter), registry.WithTags(opts.Tags))
//	keyStr := key.String()
//	picker, ok := b.store.Load(keyStr)
//	if ok {
//		return picker.(qmsbalancer.Picker).Pick(ctx, opts)
//	}
//
//	picker, err, _ := b.singleflight.Do(keyStr, func() (interface{}, error) {
//		services, err := b.resolver.Lookup(key)
//		if err != nil {
//			return nil, errors.Wrapf(ErrLookup, "err: %s", err.Error())
//		}
//		if len(services) <= 0 {
//			return nil, ErrNotFound
//		}
//		qlog.Tracef("resolver find [%d] services by [%s]", len(services), key.String())
//
//		picker := b.OnChanged(keyStr, services)
//		b.resolver.WithWatchFunc(key, WatchFunc(func(yek registry.ServiceKey, instances []*qmsbalancer.Service) {
//			b.OnChanged(keyStr, instances) //注:因为consul发现时"_""-"替换的原因，name与serviceName的值可能不完全一致
//		}))
//		return picker, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if p, ok := picker.(qmsbalancer.Picker); ok {
//		return p.(qmsbalancer.Picker).Pick(ctx, opts)
//	}
//
//	return nil, ErrArgument
//}
