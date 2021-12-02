package config

type Loader interface {
	GetConfig() *Config
	WithHandler(func(*Config))
	Stop()
}
