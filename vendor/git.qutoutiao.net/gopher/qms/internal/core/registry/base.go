package registry

import (
	"strconv"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/registryutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
)

// newRegistryBuilder 封装底层builder
func newRegistryBuilder(builder Builder) Builder {
	return &baseBuilder{
		builder,
	}
}

type baseBuilder struct {
	Builder
}

func (bb *baseBuilder) Build(config *Config) (Registry, error) {
	registry, err := bb.Builder.Build(config)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &baseRegistry{
		registry: registry,
	}, nil
}

type baseRegistry struct {
	registry Registry
}

// Register [按照服务注册规范注册服务](https://km.qutoutiao.net/pages/viewpage.action?pageId=88485252)
func (b *baseRegistry) Register(service *Service, option ...RegisterOption) (err error) {
	// 为了兼容老的 新老格式均会注册
	service.Meta[registryutil.Regtime] = strconv.FormatInt(time.Now().Unix(), 10)

	// 注册老的 paasID[-grpc]
	service.Name = service.registryName
	if err = b.registry.Register(service, option...); err != nil {
		return errors.WithStack(err)
	}

	// 注册新的 paasID[-<proto>-proto][-<env>-env] http和线上不需要对应的扩展
	service.Name = service.stdRegistryName
	if err = b.registry.Register(service, option...); err != nil {
		return errors.WithStack(err)
	}
	return
}

func (b *baseRegistry) Deregister(proto protocol.Protocol, option ...RegisterOption) {
	// 反注册老的
	//service.Name = service.getRegistryName()
	//if err := b.registry.Deregister(service, option...); err != nil {
	//	qlog.WithError(err).Errorf("deregister service[%s] failed.", service.Name)
	//}
	//
	//// 反注册新的
	//service.Name = service.getNewRegistryName()
	//if err := b.registry.Deregister(service, option...); err != nil {
	//	qlog.WithError(err).Errorf("deregister service[%s] failed.", service.Name)
	//}

	// service 实际值为rest|grpc 后续要改成真正的service，所以暂时不改
	b.registry.Deregister(proto, option...)

	return
}

// Lookup 查找后，key.name会被改为第一个查找通过的名称，此时watch只需要根据这个即可。
func (b *baseRegistry) Lookup(key *ServiceKey) (services []*Service, err error) {
	// 1. 找标准规则的
	key.name = key.stdRegistryName
	if services, err = b.registry.Lookup(key); err != nil {
		if errors.Cause(err) != ErrNotFoundService {
			return nil, errors.WithStack(err)
		}

		// 2. 找老规则的
		key.name = key.registryName
		if services, err = b.registry.Lookup(key); err != nil {
			if errors.Cause(err) != ErrNotFoundService {
				return nil, errors.WithStack(err)
			}

			// 3. 使用原始名称
			key.name = key.originName
			if services, err = b.registry.Lookup(key); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}
	return
}

func (b *baseRegistry) Watch(key *ServiceKey, watcher Watcher) {
	b.registry.Watch(key, watcher)
}

func (b *baseRegistry) Close() error {
	return b.registry.Close()
}
