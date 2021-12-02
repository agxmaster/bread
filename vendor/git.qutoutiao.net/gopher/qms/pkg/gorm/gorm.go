// pkg/gorm库
// 功能简介:
// 1.开箱即用地拿到*gorm.DB，自动从配置文件解析配置、初始化gorm，使用方不需要再次读取配置，传递参数;
// 2.自动集成了trace;
// 3.自动集成了metrics;
//
// 使用示例:
// //...
// //程序初始化阶段，判断mysql配置及网络连接的正确性.
// if err := gorm.CheckValid(); err!=nil {
//     panic(err)
// }
// //...
// mysqlName := xxx //配置中的mysql名称
// gormDB := gorm.GetClient(mysqlName)  //取到*gorm.Client
// //...            //开始使用gorm
//

package gorm

import (
	"context"
	"sync"

	gogorm "git.qutoutiao.net/golib/gorm"
	"git.qutoutiao.net/gopher/qms/pkg/conf"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var (
	nameM sync.Map
	once  sync.Once
)

type managerConfWrap struct {
	gogorm.ManagerConfig `yaml:"mysql"`
}

type (
	// Config defines config for gorm dialer
	Config = gogorm.Config
	// TODO: Client
)

// CheckValid 检验app.yaml中配置的mysql实例的配置正确性与连通性。
// 参数names是配置的实例的名称列表，如果为空，则检测所有配置的实例。
// 函数返回的error表示是否正常。
func CheckValid(names ...string) (err error) {
	initManager()

	if len(names) == 0 {
		nameM.Range(func(key, value interface{}) bool {
			if err = ping(key.(string)); err != nil {
				err = errors.Wrapf(err, "mysql(%s) is invalid", key.(string))
				return false
			}
			return true
		})
	} else {
		for _, name := range names {
			if err = ping(name); err != nil {
				return errors.Wrapf(err, "mysql(%s) is invalid", name)
			}
		}
	}
	return
}

// Client 返回某mysql实例对应的gorm官方DB对象。
// 参数names是mysql实例的名称。
// 备注: 服务初始化阶段调用CheckValid()检验过实例的有效性后，此方法访问此实例将不会再返回nil。
func GetClient(name string) *gogorm.Client {
	initManager()

	cli, err := gogorm.GetClient(name)
	if err != nil {
		qlog.WithField("name", name).Error(err)
		return nil
	}
	return cli
}

// Client 返回某mysql实例对应的gorm官方DB对象。
// Deprecated: 将于v1.3.20版本下掉
func Client(name string) *gogorm.Client {
	return GetClient(name)
}

// Client 返回某mysql实例对应的gorm官方DB对象，通过ctx参数来传递trace。
// Deprecated: 将于v1.3.20版本下掉
func ClientWithTrace(ctx context.Context, name string) *gogorm.Client {
	return Client(name)
}

// AddConfig 在app.yaml配置的基础上，新增指定mysql配置的实例。
// Deprecated: 将于v1.3.20版本下掉
func AddConfig(name string, cfg *Config) error {
	initManager()

	gogorm.Register(name, cfg)

	return ping(name)
}

func initManager() {
	once.Do(func() {
		var mconf managerConfWrap
		if err := conf.Unmarshal(&mconf); err != nil {
			qlog.Errorf("unmarshal gorm config err=%s", err)
			return
		}
		for name, config := range mconf.ManagerConfig {
			nameM.Store(name, struct{}{})

			gogorm.Register(name, config)
		}
	})
}

func ping(name string) error {
	client, err := gogorm.GetClient(name)
	if err != nil {
		return errors.WithStack(err)
	}

	mydb, err := client.DB()
	if err != nil {
		return errors.WithStack(err)
	}

	if err = mydb.Ping(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
