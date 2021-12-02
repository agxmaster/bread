package loader

import (
	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
)

type FileConfig struct {
	Filename string         `yaml:"filename"`
	Defaults *config.Config `yaml:"resty"`
}

func (config *FileConfig) NewLoader(log logger.Interface) (loader Interface, err error) {
	if config == nil {
		err = ErrNilConfig
		return
	}

	if log == nil {
		log = logger.NewWithPrefix("[FileLoader]")
	}

	return NewFileLoader(config, log)
}
