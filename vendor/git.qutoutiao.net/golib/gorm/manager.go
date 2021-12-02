package gorm

import (
	"reflect"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"git.qutoutiao.net/gopher/qulibs"
)

// Manager manages multi client instances of gorm for singleton pattern.
type Manager struct {
	single  *singleflight.Group
	clients sync.Map
	configs sync.Map
}

// NewManager creates a new manager store of gorm with configs.
func NewManager(configs ManagerConfig) *Manager {
	mgr := &Manager{
		single: new(singleflight.Group),
	}

	mgr.Load(configs)

	return mgr
}

// GetClient finds or creates a mysql client registered with the name. It uses a dummy logger for default.
func (mgr *Manager) GetClient(name string) (client *Client, err error) {
	return mgr.NewClientWithLogger(name, qulibs.NewDummyLogger())
}

// NewClientWithLogger finds or creates a mysql client registered with the name and logger given
func (mgr *Manager) NewClientWithLogger(name string, logger qulibs.Logger) (client *Client, err error) {
	if mgr == nil {
		err = ErrNotFoundClient
		return
	}

	// first, try clients store
	iface, ok := mgr.clients.Load(name)
	if ok {
		client, ok = iface.(*Client)
		if ok {
			client.SetLogger(logger)

			return
		}
	}

	// second, try creating a new client with config registered with the name.
	iface, err, _ = mgr.single.Do(name, func() (interface{}, error) {
		config, tmpErr := mgr.Config(name)
		if tmpErr != nil {
			return nil, tmpErr
		}

		// 1, create a new client
		tmpClient, tmpErr := NewWithLogger(config, logger)
		if tmpErr != nil {
			return nil, tmpErr
		}

		// 2, store the client with the name
		mgr.clients.Store(name, tmpClient)

		return tmpClient, nil
	})

	if err != nil {
		return
	}

	client, ok = iface.(*Client)

	return
}

// Config returns a config registered with the name given
func (mgr *Manager) Config(name string) (config *Config, err error) {
	if mgr == nil {
		err = ErrNotFoundConfig
		return
	}

	if len(name) == 0 {
		err = ErrInvalidConfig
		return
	}

	iface, ok := mgr.configs.Load(name)
	if !ok {
		err = ErrNotFoundConfig
		return
	}

	config, ok = iface.(*Config)
	if !ok {
		err = ErrInvalidConfig
		return
	}

	return
}

// Add registers a new config of redis with the name given.
//
// NOTE: It will remove client related to the name if existed.
func (mgr *Manager) Add(name string, config *Config) {
	if mgr == nil || len(name) == 0 || config == nil {
		return
	}

	config.FillWithDefaults()

	oldConfig, err := mgr.Config(name)
	if err != nil {
		mgr.configs.Store(name, config)

		return
	}

	if reflect.DeepEqual(oldConfig, config) {
		return
	}

	// store new config
	mgr.configs.Store(name, config)

	// remove old client
	mgr.clients.Delete(name)
}

// Del removes both client and config of redis registered with the name given.
func (mgr *Manager) Del(name string) {
	if mgr == nil {
		return
	}

	mgr.configs.Delete(name)
	mgr.clients.Delete(name)
}

// Load registers all configs with its name defined by ManagerConfig
func (mgr *Manager) Load(configs ManagerConfig) {
	if mgr == nil {
		return
	}

	for name, config := range configs {
		mgr.Add(name, config)
	}
}

func (mgr *Manager) Reload(configs ManagerConfig) error {
	if mgr == nil {
		return nil
	}

	gerr := errgroup.Group{}
	for name, config := range configs {
		gerr.Go(func() error {
			client, err := mgr.GetClient(name)
			if err != nil {
				// ignore not found
				if err == ErrNotFoundConfig {
					err = nil
				}

				return err
			}

			return client.Reload(config)
		})
	}

	return gerr.Wait()
}
