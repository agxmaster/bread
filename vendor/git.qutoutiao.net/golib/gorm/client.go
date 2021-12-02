package gorm

import (
	"context"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/logger"

	"git.qutoutiao.net/gopher/qulibs"
)

// A Client wrap *gorm.DB with best practices for development
type Client struct {
	mux    sync.RWMutex
	ctx    context.Context
	db     *gormio.DB
	value  atomic.Value
	config *Config
	logger qulibs.Logger
}

// New creates mysql client with config given and a dummy logger.
func New(config *Config) (*Client, error) {
	return NewWithLogger(config, qulibs.NewDummyLogger())
}

// NewWithLogger creates mysql client with config and logger given.
func NewWithLogger(config *Config, logger qulibs.Logger) (client *Client, err error) {
	client = &Client{
		logger: logger,
	}

	err = client.Reload(config)
	if err != nil {
		return nil, err
	}

	return
}

func (c *Client) load() *Client {
	db, ok := c.value.Load().(*gormio.DB)
	if ok {
		return &Client{
			db:     db.WithContext(context.Background()),
			value:  atomic.Value{},
			config: c.config,
			logger: c.logger,
		}
	}

	return c
}

func (c *Client) SetLogger(logger qulibs.Logger) {
	if c == nil || logger == nil {
		return
	}

	c.mux.Lock()
	c.logger = logger
	c.mux.Unlock()
}

// Reload refreshes internal gorm.DB with config given.
func (c *Client) Reload(config *Config) (err error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	config.FillWithDefaults()

	mycfg, err := config.NewMycfg()
	if err != nil {
		return
	}

	if c.config != nil && reflect.DeepEqual(c.config.mycfg, mycfg) {
		return
	}

	var dialer gormio.Dialector
	switch config.Driver {
	case "mysql":
		dialer = mysql.Open(mycfg.FormatDSN())

	case "postgres":
		dialer = postgres.Open(mycfg.FormatDSN())

	case "sqlite":
		dialer = sqlite.Open(mycfg.FormatDSN())
	}

	logcfg := logger.Config{
		SlowThreshold: 100 * time.Millisecond,
		Colorful:      false,
		LogLevel:      logger.Warn,
	}
	if config.DebugSQL {
		logcfg.LogLevel = logger.Info
	}

	var logface logger.Interface
	if w, ok := c.logger.(logger.Writer); ok {
		logface = logger.New(w, logcfg)
	} else {
		logface = logger.New(log.New(os.Stderr, "\r\n", log.LstdFlags), logcfg)
	}

	db, err := gormio.Open(dialer, &gormio.Config{
		Logger:               logface,
		PrepareStmt:          true,
		DisableAutomaticPing: false,
		AllowGlobalUpdate:    false,
	})
	if err != nil {
		return
	}

	mydb, err := db.DB()
	if err != nil {
		return
	}

	err = mydb.Ping()
	if err != nil {
		return
	}

	if config.MaxOpenConns > 0 {
		mydb.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		mydb.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxLifetime > 0 {
		mydb.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Second)
	}

	// NOTE: It could cause runtime error for running client!
	if oldDB, ok := c.value.Load().(*gormio.DB); ok {
		defer func(old *gormio.DB, dsn string) {
			oldMydb, err := old.DB()
			if err == nil {
				qulibs.Errorf("%T.Close(%s): %+v", old, dsn, oldMydb.Close())
			}
		}(oldDB, c.config.mycfg.FormatDSN())
	}

	c.value.Store(db)

	c.config = config
	c.config.mycfg = mycfg

	registerTraceCallbacks(c)
	registerMetricsCallbacks(c)

	return
}

// SelectDB switches to a new database of dbname given by creating a new gorm instance.
func (c *Client) SelectDB(dbname string) (client *Client, err error) {
	c.mux.RLock()
	if c.config.IsEqualDB(dbname) {
		c.mux.RUnlock()

		return c, nil
	}

	config, err := c.config.NewWithDB(dbname)
	if err != nil {
		c.mux.RUnlock()
		return
	}

	name := config.Name()

	// first, try loading a client from default manager
	client, err = singleton.NewClientWithLogger(name, c.logger)
	if err == nil {
		c.mux.RUnlock()

		return client, nil
	}

	c.mux.RUnlock()

	// second, register new client for default manager
	c.mux.Lock()
	defer c.mux.Unlock()

	singleton.Add(name, config)

	return singleton.NewClientWithLogger(name, c.logger)
}
