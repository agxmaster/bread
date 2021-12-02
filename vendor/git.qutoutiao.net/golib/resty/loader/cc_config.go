package loader

import (
	"git.qutoutiao.net/gopher/cc-client-go"
	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
)

type CcConfig struct {
	Project  string         `yaml:"project"`
	Token    string         `yaml:"token"`
	Key      string         `yaml:"key"`
	Env      cc.Env         `yaml:"env"`
	Defaults *config.Config `yaml:"resty"`
}

func (config *CcConfig) NewLoader(log logger.Interface) (loader Interface, err error) {
	if config == nil {
		err = ErrNilConfig
		return
	}

	if log == nil {
		log = logger.NewWithPrefix("[CCLoader]")
	}

	return NewCcLoader(config, log)
}
