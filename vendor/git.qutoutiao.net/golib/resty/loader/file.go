package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"

	"git.qutoutiao.net/golib/resty/config"
	"git.qutoutiao.net/golib/resty/logger"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type fileLoader struct {
	filename string
	handler  func(config *config.Config)
	value    atomic.Value
	watcher  *fsnotify.Watcher
	log      logger.Interface
}

func NewFileLoader(cfg *FileConfig, log logger.Interface) (*fileLoader, error) {
	if cfg == nil || cfg.Filename == "" {
		return nil, fmt.Errorf("invalid config")
	}

	loader := &fileLoader{
		filename: cfg.Filename,
		value:    atomic.Value{},
		log:      log,
	}

	err := loader.init(cfg.Defaults)
	if err != nil {
		return nil, err
	}

	go loader.watch()

	return loader, nil
}

func (loader *fileLoader) WithHandler(fn func(*config.Config)) {
	if loader == nil {
		return
	}

	loader.handler = fn
}

// GetConfig returns config loaded from remote
func (loader *fileLoader) GetConfig() *config.Config {
	if loader == nil {
		return defaultConfig
	}

	iface := loader.value.Load()
	if iface == nil {
		return nil
	}

	cfg, ok := iface.(*config.Config)
	if !ok {
		return nil
	}

	return cfg
}

func (loader *fileLoader) Stop() {
	if loader == nil {
		return
	}

	if loader.watcher == nil {
		return
	}

	loader.watcher.Close()
	loader.watcher = nil
}

func (loader *fileLoader) init(defaults *config.Config) error {
	data, err := ioutil.ReadFile(loader.filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if defaults == nil {
			return err
		}

		value, err := yaml.Marshal(defaults)
		if err != nil {
			return fmt.Errorf("encode config(%s): %v", loader.filename, err)
		}

		err = ioutil.WriteFile(loader.filename, value, 0644)
		if err != nil {
			return fmt.Errorf("init config(%s): %v", loader.filename, err)
		}

		loader.value.Store(defaults)
		return nil
	}

	cfg := new(config.Config)

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", loader.filename, cfg, err)
		return err
	}

	loader.value.Store(cfg)

	return nil
}

func (loader *fileLoader) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(loader.filename)
	if err != nil {
		return err
	}

	loader.watcher = watcher

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			loader.log.Debugf("watch(%s): %s", loader.filename, event.String())
			switch {
			case event.Op&fsnotify.Write == fsnotify.Write,
				event.Op&fsnotify.Chmod == fsnotify.Chmod:
				data, err := ioutil.ReadFile(loader.filename)
				if err == nil {
					loader.watchHandler(event.Name, data)
				} else {
					loader.log.Errorf("watch ioutil.ReadFile(%s): %v", loader.filename, err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}

			loader.log.Errorf("watch(%s): %v", loader.filename, err)
		}
	}
}

func (loader *fileLoader) watchHandler(event string, data []byte) {
	cfg := new(config.Config)

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		loader.log.Errorf("yaml.Unmarshal(%s, %T): %v", event, cfg, err)
		return
	}

	loader.value.Store(cfg)

	if loader.handler != nil {
		loader.handler(cfg)
	}
}
