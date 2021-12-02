package redis

import (
	"context"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	redisv8 "github.com/go-redis/redis/v8"

	"git.qutoutiao.net/golib/redis/metrics"
	"git.qutoutiao.net/gopher/qulibs"
	"github.com/prometheus/client_golang/prometheus"
)

// A Client wraps redis client with custom features
type Client struct {
	mux    sync.RWMutex
	ctx    context.Context
	config *Config
	value  atomic.Value
	logger qulibs.Logger
}

// New creates a new redis value with config given and a dummy logger.
func New(config *Config) (*Client, error) {
	return NewWithLogger(config, qulibs.NewDummyLogger())
}

// NewWithLogger creates a new redis value with config and logger given.
func NewWithLogger(config *Config, logger qulibs.Logger) (client *Client, err error) {
	client = &Client{
		ctx:    context.Background(),
		logger: logger,
	}

	err = client.Reload(config)
	if err != nil {
		return nil, err
	}

	go client.metrics(client.ctx)

	return
}

func (c *Client) load() *redisv8.Client {
	// for dummy case
	if c == nil {
		return nil
	}

	client, ok := c.value.Load().(*redisv8.Client)
	if !ok {
		return nil
	}

	return client
}

func (c *Client) SetLogger(logger qulibs.Logger) {
	if c == nil || logger == nil {
		return
	}

	c.mux.Lock()
	c.logger = logger
	c.mux.Unlock()
}

func (c *Client) WithContext(ctx context.Context) (client *Client) {
	if ctx == nil {
		ctx = context.Background()
	}

	clone := &Client{
		ctx:    ctx,
		config: c.config,
		value:  atomic.Value{},
		logger: c.logger,
	}
	clone.value.Store(c.load().WithContext(ctx))

	return clone
}

func (c *Client) WithTimeout(duration time.Duration) (client *Client) {
	clone := &Client{
		ctx:    c.ctx,
		config: c.config,
		value:  atomic.Value{},
		logger: c.logger,
	}
	clone.value.Store(c.load().WithTimeout(duration))

	return clone
}

// SelectDB changes db by coping out a new value.
//
// NOTE: There maybe a deadlock if internal invocations panic!!!
func (c *Client) SelectDB(db int) (*Client, error) {
	c.mux.RLock()

	opts := c.load().Options()
	if opts.DB == db {
		c.mux.RUnlock()

		return c, nil
	}

	c.mux.RUnlock()

	// creates a new value
	c.mux.Lock()
	defer c.mux.Unlock()

	config := &Config{
		Network:      opts.Network,
		Addr:         opts.Addr,
		Passwd:       opts.Password,
		DB:           db,
		DialTimeout:  int(opts.DialTimeout / time.Millisecond),
		ReadTimeout:  int(opts.ReadTimeout / time.Millisecond),
		WriteTimeout: int(opts.WriteTimeout / time.Millisecond),
		PoolSize:     opts.PoolSize,
		PoolTimeout:  int(opts.PoolTimeout / time.Millisecond),
		MinIdleConns: opts.MinIdleConns,
		MaxRetries:   opts.MaxRetries,
	}

	name := config.Name()

	// first, try loading a value from default manager
	client, err := singleton.NewClientWithLogger(name, c.logger)
	if err == nil {
		return client, nil
	}
	c.logger.Warnf("singleton.GetClient(%s): %v", name, err)

	// second, register new value with default manager
	singleton.Add(name, config)

	return singleton.NewClientWithLogger(name, c.logger)
}

// Reload refreshes redis client with new config.
func (c *Client) Reload(config *Config) (err error) {
	if c == nil {
		err = ErrNotFoundClient
		return
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	config.FillWithDefaults()

	if reflect.DeepEqual(c.config, config) {
		return
	}

	client := redisv8.NewClient(&redisv8.Options{
		Network:      config.Network,
		Addr:         config.Addr,
		Password:     config.Passwd,
		DB:           config.DB,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Millisecond,
		PoolSize:     config.PoolSize,
		PoolTimeout:  time.Duration(config.PoolTimeout) * time.Millisecond,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
	})
	client.AddHook(OpentracingHook{
		db:              strconv.Itoa(config.DB),
		addr:            config.Addr,
		includeNotFound: config.TraceIncludeNotFound,
	})

	// NOTE: It could cause runtime error for running client!
	if old := c.load(); old != nil {
		defer func(oldClient *redisv8.Client) {
			err := oldClient.Close()
			if err != nil {
				qulibs.Errorf("redis.Close(%+v): %v", oldClient.Options(), err)
			}
		}(old)
	}

	c.value.Store(client)
	c.config = config

	return
}

func (c *Client) cmdMetrics(cmd redisv8.Cmder, issuedAt time.Time) prometheus.Labels {
	labels := prometheus.Labels{"cmd": cmd.Name(), "to": c.config.Addr, "status": redisSuccess}
	if err := cmd.Err(); err != nil && (!c.config.MetricsIncludeNotFound || err != Nil) {
		labels["status"] = redisFailed
	}

	metrics.ObserveCmd(labels, issuedAt)
	metrics.IncCmd(labels)
	return labels
}
