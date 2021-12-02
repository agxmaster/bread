package gorm

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config defines config for gorm dialer
type Config struct {
	Driver                 string        `yaml:"driver"`                    // 数据库类型，当前支持：mysql，postgres，sqlite
	DSN                    string        `yaml:"dsn"`                       // 数据库 DSN
	DialTimeout            time.Duration `yaml:"dial_timeout"`              // 连接超时时间，默认 1000ms
	ReadTimeout            time.Duration `yaml:"read_timeout"`              // socket 读超时时间，默认 3000ms
	WriteTimeout           time.Duration `yaml:"write_timeout"`             // socket 写超时时间，默认 3000ms
	MaxOpenConns           int           `yaml:"max_open_conns"`            // 最大连接数，默认 200
	MaxIdleConns           int           `yaml:"max_idle_conns"`            // 最大空闲连接数，默认 80
	MaxLifetime            int           `yaml:"max_life_time"`             // 空闲连接最大存活时间，默认 600s
	TraceIncludeNotFound   bool          `yaml:"trace_include_not_found"`   // 是否将NotFound error作为错误记录在trace中，默认为否
	MetricsIncludeNotFound bool          `yaml:"metrics_include_not_found"` // 是否将NotFound error作为错误记录在metrics中，默认为否
	DebugSQL               bool          `yaml:"debug_sql"`                 // 开启 SQL 调试模式，即输出所有 SQL 语句

	// internal
	mycfg *mysql.Config `yaml:"-"`
}

// Name returns name of the gorm dialer for Manager.
func (c *Config) Name() string {
	dsn, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return c.Driver
	}

	return fmt.Sprintf("%s(%s/%s)", c.Driver, dsn.Addr, dsn.DBName)
}

// FillWithDefaults apply default values for field with invalid db.
func (c *Config) FillWithDefaults() {
	if c == nil {
		return
	}

	if c.Driver == "" {
		c.Driver = DefaultDriver
	}

	if c.DialTimeout <= 0 {
		c.DialTimeout = MaxDialTimeout
	}

	if c.ReadTimeout <= 0 {
		c.ReadTimeout = MaxReadTimeout
	}

	if c.WriteTimeout <= 0 {
		c.WriteTimeout = MaxWriteTimeout
	}

	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = MaxOpenConn
	}

	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = MaxIdleConn
	}

	if c.MaxLifetime <= 0 {
		c.MaxLifetime = MaxLifetime
	}
}

// NewMycfg returns a *mysql.Config with timeout settings
func (c *Config) NewMycfg() (dsn *mysql.Config, err error) {
	dsn, err = mysql.ParseDSN(c.DSN)
	if err != nil {
		return
	}

	// adjust timeout of DSN
	if dsn.Timeout <= 0 {
		dsn.Timeout = c.DialTimeout * time.Millisecond
	}
	if dsn.ReadTimeout <= 0 {
		dsn.ReadTimeout = c.ReadTimeout * time.Millisecond
	}
	if dsn.WriteTimeout <= 0 {
		dsn.WriteTimeout = c.WriteTimeout * time.Millisecond
	}

	// sync
	c.DSN = dsn.FormatDSN()

	return
}

// NewWithDB creates a new config with the database name given for gorm dialer
func (c *Config) NewWithDB(dbname string) (*Config, error) {
	mycfg, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return nil, err
	}

	mycfg.DBName = dbname

	copied := *c
	copied.DSN = mycfg.FormatDSN()
	copied.mycfg = mycfg

	return &copied, nil
}

// IsEqualDB returns true if database specified by dsn is equal to dbname given.
func (c *Config) IsEqualDB(dbname string) bool {
	dsn, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return false
	}

	return strings.Compare(dsn.DBName, dbname) == 0
}

// A ManagerConfig defines a list of gorm dialer config with its name
type ManagerConfig map[string]*Config
