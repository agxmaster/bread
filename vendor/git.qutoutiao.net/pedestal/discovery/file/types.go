package file

import (
	"git.qutoutiao.net/pedestal/discovery/registry"
)

type Loader interface {
	Load(registry.ServiceKey) ([]*registry.Service, error)
}
