package redis

import (
	"fmt"
	"runtime"
)

// A config of go redis
type Config struct {
	Network              string `yaml:"network"`                 //网络类型，支持：tcp，unix；默认tcp
	Addr                 string `yaml:"addr"`                    //网络地址，ip:port，如：172.0.0.1:6379
	Passwd               string `yaml:"password"`                //密码
	DB                   int    `yaml:"database"`                //redis database，默认0；当前已不推荐使用多DB，该配置只为兼容一些存量系统多DB的使用
	DialTimeout          int    `yaml:"dial_timeout"`            //连接超时时间，默认1000ms
	ReadTimeout          int    `yaml:"read_timeout"`            //socket 读超时时间，默认100ms
	WriteTimeout         int    `yaml:"write_timeout"`           //socket 写超时时间，默认100ms
	PoolSize             int    `yaml:"pool_size"`               //连接池最大数量，默认200
	PoolTimeout          int    `yaml:"pool_timeout"`            //从连接池获取连接超时时间，默认ReadTimeout + 1000ms
	MinIdleConns         int    `yaml:"min_idle_conns"`          //连接池最小空闲连接数，默认30
	MaxRetries           int    `yaml:"max_retries"`             //重试次数，默认0
	TraceIncludeNotFound bool   `yaml:"trace_include_not_found"` //是否将key NotFound error作为错误记录在trace中，默认为否
	MetricsIncludeNotFound bool   `yaml:"metrics_include_not_found"`
}

// Name returns value name of the config
func (c *Config) Name() string {
	return fmt.Sprintf("%s(%s/%d)", c.Network, c.Addr, c.DB)
}

// FillWithDefaults apply default values for fields with invalid values.
func (c *Config) FillWithDefaults() {
	maxCPU := runtime.NumCPU()

	if c.DialTimeout <= 0 || c.DialTimeout > MaxDialTimeout*maxCPU {
		c.DialTimeout = MaxDialTimeout
	}

	if c.ReadTimeout <= 0 || c.ReadTimeout > MaxReadTimeout*maxCPU {
		c.ReadTimeout = MaxReadTimeout
	}

	if c.WriteTimeout <= 0 || c.WriteTimeout > MaxWriteTimeout*maxCPU {
		c.WriteTimeout = MaxWriteTimeout
	}

	if c.PoolSize <= 0 || c.PoolSize > MaxPoolSize*maxCPU {
		c.PoolSize = MaxPoolSize
	}

	if c.PoolTimeout <= 0 || c.PoolTimeout > MaxPoolTimeout*maxCPU {
		c.PoolTimeout = MaxPoolTimeout
	}

	if c.MinIdleConns <= 0 || c.MinIdleConns > MinIdleConns*maxCPU {
		c.MinIdleConns = MinIdleConns
	}

	if c.MaxRetries < 0 || c.MaxRetries > MaxRetries*maxCPU {
		c.MaxRetries = MaxRetries
	}
}

// A ManagerConfig defines a list of redis config with its name
type ManagerConfig map[string]*Config
