package loader

import (
	"fmt"
	"io"
	"sync/atomic"

	"git.qutoutiao.net/gopher/cc-client-go"
	center "git.qutoutiao.net/gopher/cc-client-go/config"
	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
	"gopkg.in/yaml.v3"
)

type ccLoader struct {
	client  *cc.ConfigCenter
	key     string
	handler func(*config.Config)
	value   atomic.Value
	closer  io.Closer
	log     logger.Interface
}

func NewCcLoader(cfg *CcConfig, log logger.Interface) (*ccLoader, error) {
	if cfg == nil || cfg.Project == "" || cfg.Token == "" || cfg.Key == "" || cfg.Env == "" {
		return nil, fmt.Errorf("invalid config")
	}

	loader := &ccLoader{
		key:   cfg.Key,
		value: atomic.Value{},
		log:   log,
	}

	client, closer, err := center.NewConfiguration(cfg.Project, cfg.Token, cfg.Env).
		NewCC(center.OnChange(func(center *cc.ConfigCenter) error {
			data := center.GetString(cfg.Key, "")
			if len(data) == 0 {
				return fmt.Errorf("config(key=%s) is empty", cfg.Key)
			}

			loader.watchHandler(data)

			return nil
		}))
	if err != nil && err != cc.ErrBackup {
		return nil, err
	}

	loader.client = client
	loader.closer = closer

	err = loader.init(cfg.Defaults)
	if err != nil {
		return nil, err
	}

	go loader.watch()

	return loader, nil
}

func (loader *ccLoader) WithHandler(fn func(*config.Config)) {
	if loader == nil {
		return
	}

	loader.handler = fn
}

// GetConfig returns config loaded from remote
func (loader *ccLoader) GetConfig() *config.Config {
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

func (loader *ccLoader) Stop() {
	if loader == nil {
		return
	}

	if loader.closer == nil {
		return
	}

	loader.closer.Close()
	loader.closer = nil
}

func (loader *ccLoader) init(defaults *config.Config) error {
	data := loader.client.GetString(loader.key, "")
	if len(data) == 0 {
		if defaults == nil {
			return fmt.Errorf("config not found")
		}

		loader.value.Store(defaults)
		return nil
	}

	cfg := new(config.Config)

	err := yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", loader.key, cfg, err)
		return err
	}

	loader.value.Store(cfg)

	return nil
}

func (loader *ccLoader) watch() error {
	return nil
}

func (loader *ccLoader) watchHandler(data string) {
	cfg := new(config.Config)

	err := yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", loader.key, cfg, err)
		return
	}

	loader.value.Store(cfg)

	if loader.handler != nil {
		loader.handler(cfg)
	}
}
