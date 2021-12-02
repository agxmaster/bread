package consul

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/consul/api"

	"git.qutoutiao.net/pedestal/discovery/cache/file"
	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

const (
	FormatHealthAPI  = "health"
	FormatCatalogAPI = "catalog"
)

// Cache implements cache.Interface by wrapping file.Cache with cache.FormatConsul support.
type Cache struct {
	format string
	*file.Cache
}

func New(root string) *Cache {
	return NewWithFormat(root, FormatHealthAPI)
}

func NewWithFormat(root, format string) *Cache {
	fd := file.New(root)

	return &Cache{
		format: format,
		Cache:  fd,
	}
}

// Load tries to parse services for the key from local cached file.
// NOTE: it parses data dumped from http://consul/v1/health/service/<service> api by overwriting file.Cache.Load implementation.
func (c *Cache) Load(key registry.ServiceKey) ([]*registry.Service, error) {
	switch c.format {
	case FormatHealthAPI:
		return c.loadHealthAPI(key)

	case FormatCatalogAPI:
		return c.loadCatalogAPI(key)
	}

	return nil, fmt.Errorf("invalid consul api format of %s, available values are [%s|%s]", c.format, FormatHealthAPI, FormatCatalogAPI)
}

func (c *Cache) loadHealthAPI(key registry.ServiceKey) ([]*registry.Service, error) {
	filename := c.Filename(key)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(errors.ErrNotFound)
		}

		return nil, err
	}

	// try to parse consul entries
	var entries []*api.ServiceEntry

	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	// build services for registry
	services := make([]*registry.Service, len(entries))
	for i, entry := range entries {
		if entry.Service == nil {
			continue
		}

		services[i] = &registry.Service{
			ID:   entry.Service.ID,
			Name: key.Name,
			IP:   entry.Service.Address,
			Port: entry.Service.Port,
			Tags: entry.Service.Tags,
			Meta: entry.Service.Meta,
		}
		if entry.Service.Weights.Passing > 0 {
			services[i].Weight = int32(entry.Service.Weights.Passing)
		}
	}

	if len(services) == 0 {
		return nil, errors.Wrap(errors.ErrNotFound)
	}

	return services, nil
}

func (c *Cache) loadCatalogAPI(key registry.ServiceKey) ([]*registry.Service, error) {
	filename := c.Filename(key)

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(errors.ErrNotFound)
		}

		return nil, err
	}

	// try to parse consul entries
	var entries []*api.CatalogService

	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, err
	}

	// build services for registry
	services := make([]*registry.Service, len(entries))
	for i, entry := range entries {
		services[i] = &registry.Service{
			ID:   entry.ServiceID,
			Name: key.Name,
			IP:   entry.ServiceAddress,
			Port: entry.ServicePort,
			Tags: entry.ServiceTags,
			Meta: entry.ServiceMeta,
		}
		if entry.ServiceWeights.Passing > 0 {
			services[i].Weight = int32(entry.ServiceWeights.Passing)
		}
	}

	if len(services) == 0 {
		return nil, errors.Wrap(errors.ErrNotFound)
	}

	return services, nil
}
