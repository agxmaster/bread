package loader

import (
	"fmt"
	stdlog "log"
	"os"
	"sync/atomic"

	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"gopkg.in/yaml.v3"
)

type consulLoader struct {
	client  *api.Client
	key     string
	handler atomic.Value
	value   atomic.Value
	plan    *watch.Plan
	log     logger.Interface
}

func NewConsulLoader(cfg *ConsulConfig, log logger.Interface) (*consulLoader, error) {
	if cfg == nil || cfg.Key == "" || cfg.Addr == "" {
		return nil, fmt.Errorf("invalid config")
	}

	config := api.DefaultConfig()
	config.Address = cfg.Addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	loader := &consulLoader{
		client: client,
		value:  atomic.Value{},
		key:    cfg.Key,
		log:    log,
	}

	err = loader.init(cfg.Defaults)
	if err != nil {
		return nil, err
	}

	go loader.watch()

	return loader, nil
}

func (loader *consulLoader) WithHandler(fn func(*config.Config)) {
	if loader == nil {
		return
	}

	loader.handler.Store(fn)
}

// GetConfig returns config loaded from remote
func (loader *consulLoader) GetConfig() *config.Config {
	if loader == nil {
		return defaultConfig
	}

	iface := loader.value.Load()
	if iface == nil {
		return nil
	}

	cfg, ok := iface.(*config.Config)
	if !ok {
		return nil
	}

	return cfg
}

func (loader *consulLoader) Stop() {
	if loader == nil {
		return
	}

	if loader.plan == nil {
		return
	}

	loader.plan.Stop()
}

func (loader *consulLoader) init(defaults *config.Config) error {
	kv, _, err := loader.client.KV().Get(loader.key, &api.QueryOptions{
		AllowStale: true,
	})
	if err != nil {
		return err
	}

	if kv == nil {
		if defaults == nil {
			return fmt.Errorf("config not found")
		}

		value, err := yaml.Marshal(defaults)
		if err != nil {
			return fmt.Errorf("encode config(%s): %v", loader.key, err)
		}

		kv = &api.KVPair{
			Key:   loader.key,
			Value: value,
		}

		_, err = loader.client.KV().Put(kv, &api.WriteOptions{})
		if err != nil {
			return fmt.Errorf("init config(%s): %v", loader.key, err)
		}

		loader.value.Store(defaults)

		return nil
	}

	cfg := new(config.Config)

	err = yaml.Unmarshal(kv.Value, &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", kv.Key, cfg, err)
		return err
	}

	loader.value.Store(cfg)

	return nil
}

func (loader *consulLoader) watch() error {
	plan, err := watch.Parse(map[string]interface{}{
		"stale": true,
		"type":  "key",
		"key":   loader.key,
	})
	if err != nil {
		return err
	}

	plan.Handler = loader.watchHandler

	if err := plan.RunWithClientAndLogger(loader.client, stdlog.New(os.Stderr, "[Consul Interface]", stdlog.LstdFlags)); err != nil {
		return err
	}

	loader.plan = plan

	return nil
}

func (loader *consulLoader) watchHandler(idx uint64, data interface{}) {
	pair, ok := data.(*api.KVPair)
	if !ok {
		return
	}
	if pair == nil {
		return
	}
	loader.log.Debugf("received new config of %v(%d)", pair.Key, pair.ModifyIndex)

	cfg := new(config.Config)

	err := yaml.Unmarshal(pair.Value, &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", pair.Key, cfg, err)
		return
	}

	loader.value.Store(cfg)

	handler, ok := loader.handler.Load().(func(*config.Config))
	if ok {
		handler(cfg)
	}
}
