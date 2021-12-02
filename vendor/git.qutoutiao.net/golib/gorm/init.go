package gorm

import (
	"git.qutoutiao.net/gopher/qulibs"
	"git.qutoutiao.net/gopher/qulibs/config/file"
)

const ComponentName = "mysql"

var (
	singleton *Manager
)

func init() {
	singleton = NewManager(nil)
}

// GetClient returns *Client for given name registered in singleton
func GetClient(name string) (*Client, error) {
	return singleton.GetClient(name)
}

// Init initializes gorm clients by parsing config from filename.
func Init(filename string) error {
	loader := file.New(filename)

	var config ManagerConfig

	err := loader.Load(ComponentName, &config)
	if err != nil {
		qulibs.Errorf("%T.Load(%s, %s, %T): %+v", loader, ComponentName, filename, config, err)

		return err
	}

	singleton.Load(config)
	return nil
}

// Load returns *Client related with the name given.
func Load(name string) (*Client, error) {
	return singleton.GetClient(name)
}

// Register adds new config with the name, it will overwrite the client of the same name already registered.
func Register(name string, config *Config) {
	singleton.Add(name, config)
}
