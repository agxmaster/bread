// pkg/redis库
// 功能简介:
// 1.开箱即用地拿到go-redis的Client，自动从配置文件解析配置初始化go-redis，使用方不需要再次读取配置，传递参数;
// 2.自动集成了trace;
// 3.自动集成了metrics;
//
// 使用示例:
// //...
// //程序初始化阶段，判断redis配置及网络连接的正确性.
// if err := redis.CheckValid(); err!=nil {
//     panic(err)
// }
// //...
// redisName := xxx //配置中的redis名称
// redisCli := redis.GetClient(redisName)  //取到go-redis的Client对象
// //...            //开始使用go-redis
//

package redis

import (
	"context"
	"sync"

	goredis "git.qutoutiao.net/golib/redis"
	"git.qutoutiao.net/gopher/qms/pkg/conf"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var (
	nameM sync.Map
	once  sync.Once
)

type (
	// A config of go redis
	Config = goredis.Config
	// TODO: Client
)

type managerConfWrap struct {
	goredis.ManagerConfig `yaml:"redis"`
}

// CheckValid 检验app.yaml中配置的redis实例的配置正确性与连通性。
// 参数names是配置的实例的名称列表，如果为空，则检测所有配置的实例。
// 函数返回的error表示是否正常。
func CheckValid(names ...string) (err error) {
	initManager()

	if len(names) == 0 {
		nameM.Range(func(key, value interface{}) bool {
			if err = ping(key.(string)); err != nil {
				err = errors.Wrapf(err, "redis(%s) is invalid", key.(string))
				return false
			}
			return true
		})
	} else {
		for _, name := range names {
			if err = ping(name); err != nil {
				return errors.Wrapf(err, "redis(%s) is invalid", name)
			}
		}
	}
	return nil
}

// GetClient 返回某redis实例对应的Client对象。
// 参数names是redis实例的名称。
// 备注: 服务初始化阶段调用CheckValid()检验过实例的有效性后，此方法访问此实例将不会再返回nil。
func GetClient(name string) *goredis.Client {
	initManager()

	client, err := goredis.GetClient(name)
	if err != nil {
		qlog.WithField("name", name).Error(err)
		return nil
	}
	return client
}

// Client 返回某redis实例对应的Client对象。
// Deprecated: 将于v1.3.20版本下掉
func Client(name string) *goredis.Client {
	return GetClient(name)
}

// Client 返回某redis实例对应的go-redis官方Client对象，通过ctx参数来传递trace。
// Deprecated: 将于v1.3.20版本下掉
func ClientWithTrace(ctx context.Context, name string) *goredis.Client {
	return Client(name)
}

// GetConfig 返回某redis示例的配置信息。
// Deprecated: 将于v1.3.20版本下掉
func GetConfig(name string) *Config {
	iface, ok := nameM.Load(name)
	if ok {
		return iface.(*Config)
	}
	return nil
}

func initManager() {
	once.Do(func() {
		var mconf managerConfWrap
		if err := conf.Unmarshal(&mconf); err != nil {
			qlog.Errorf("unmarshal gorm config err=%s", err)
			return
		}
		for name, config := range mconf.ManagerConfig {
			nameM.Store(name, config)

			goredis.Register(name, config)
		}
	})
}

func ping(name string) error {
	client, err := goredis.GetClient(name)
	if err != nil {
		return errors.WithStack(err)
	}

	if _, err = client.Ping(context.Background()).Result(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
