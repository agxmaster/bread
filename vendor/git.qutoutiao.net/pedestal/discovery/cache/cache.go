package cache

import (
	"git.qutoutiao.net/pedestal/discovery/cache/consul"
	"git.qutoutiao.net/pedestal/discovery/cache/file"
	"git.qutoutiao.net/pedestal/discovery/errors"
)

func New(opts ...Option) (Interface, error) {
	o := new(options)
	for _, opt := range opts {
		opt(o)
	}

	if !o.format.IsValid() {
		return nil, errors.ErrInvalidCache
	}

	var c Interface
	switch o.format {
	case FormatConsul:
		c = consul.New(o.root)

	case FormatConsulHealth:
		c = consul.NewWithFormat(o.root, consul.FormatHealthAPI)

	case FormatConsulCatalog:
		c = consul.NewWithFormat(o.root, consul.FormatCatalogAPI)

	case FormatDiscovery:
		c = file.New(o.root)
	}

	return c, nil
}
