package file

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

// Cache implements cache.Interface of local file with cache.FormatDiscovery support.
type Cache struct {
	root    string
	once    sync.Once
	onceErr error
}

func New(root string) *Cache {
	return &Cache{
		root: root,
	}
}

// Filename returns filename of cached file for the key given.
func (c *Cache) Filename(key registry.ServiceKey) string {
	c.once.Do(func() {
		c.onceErr = os.MkdirAll(c.root, 0777)
		if c.onceErr != nil {
			logger.Errorf("os.MkdirAll(%s): %+v", c.root, c.onceErr)
		}
	})

	return filepath.Join(c.root, key.ToString())
}

// LastModify tries to resolve ctime of cached file for the key given.
func (c *Cache) LastModify(key registry.ServiceKey) (time.Time, error) {
	filename := c.Filename(key)
	if c.onceErr != nil {
		return time.Time{}, c.onceErr
	}

	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, errors.ErrNotFound
		}

		// avoid flush filename by returning now forever
		return time.Now(), err
	}

	return info.ModTime(), nil
}

// Store tries to persist services for the key within local cached file.
func (c *Cache) Store(key registry.ServiceKey, services interface{}) error {
	filename := c.Filename(key)
	if c.onceErr != nil {
		return c.onceErr
	}

	data, err := json.Marshal(services)
	if err != nil {
		return err
	}

	//may slow and safe write file
	return WriteAtomicWithPerms(filename, data, 0666)
}

// Load tries to parse services for the key from local cached file.
func (c *Cache) Load(key registry.ServiceKey) ([]*registry.Service, error) {
	filename := c.Filename(key)
	if c.onceErr != nil {
		return nil, c.onceErr
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(errors.ErrNotFound)
		}

		return nil, err
	}

	var services []*registry.Service

	err = json.Unmarshal(data, &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}
