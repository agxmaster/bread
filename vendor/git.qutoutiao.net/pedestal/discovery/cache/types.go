package cache

import (
	"time"

	"git.qutoutiao.net/pedestal/discovery/registry"
)

// cache interface.
type Interface interface {
	LastModify(registry.ServiceKey) (time.Time, error)
	Store(registry.ServiceKey, interface{}) error
	Load(registry.ServiceKey) ([]*registry.Service, error)
}
