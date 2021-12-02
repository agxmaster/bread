package loader

import (
	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
)

type ConsulConfig struct {
	Addr     string         `yaml:"addr"`
	Key      string         `yaml:"key"`
	Defaults *config.Config `yaml:"resty"`
}

func (config *ConsulConfig) NewLoader(log logger.Interface) (loader Interface, err error) {
	if config == nil {
		err = ErrNilConfig
		return
	}

	if log == nil {
		log = logger.NewWithPrefix("[ConsulLoader]")
	}

	return NewConsulLoader(config, log)
}
