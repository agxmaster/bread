package loader

import (
	"errors"

	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
)

var (
	defaultConfig = &config.Config{}
)

type Interface interface {
	GetConfig() *config.Config
	WithHandler(func(*config.Config))
	Stop()
}

func New(cfg *config.LoaderConfig) (iface Interface, err error) {
	if !cfg.IsValid() {
		err = errors.New("invalid loader config")
		return
	}

	switch cfg.Provider {
	case "file":
		var fileConfig *FileConfig

		err = cfg.UnmarshalYAML(&fileConfig)
		if err != nil {
			return
		}

		clog := logger.NewWithPrefix("File")

		return NewFileLoader(fileConfig, clog)

	case "consul":
		var consulConfig *ConsulConfig

		err = cfg.UnmarshalYAML(&consulConfig)
		if err != nil {
			return
		}

		clog := logger.NewWithPrefix("Consul")

		return NewConsulLoader(consulConfig, clog)

	case "cc":
		var ccConfig *CcConfig

		err = cfg.UnmarshalYAML(&ccConfig)
		if err != nil {
			return
		}

		clog := logger.NewWithPrefix("CC")

		return NewCcLoader(ccConfig, clog)
	}

	err = errors.New("unsupported loader, available values are [file|consul|cc]")
	return
}
