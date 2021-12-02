package file

import (
	"io/ioutil"

	"git.qutoutiao.net/gopher/qulibs"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	filename string
}

func New(filename string) *Config {
	return &Config{filename: filename}
}

func (c *Config) Load(component string, v interface{}) error {
	data, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return errors.Wrapf(err, "ioutil.ReadFile(%s)", c.filename)
	}

	// try parse as components format
	var kv map[string]interface{}

	err = yaml.Unmarshal(data, &kv)
	if err != nil {
		qulibs.Warnf("cannot parse config as components format, yaml.Unmarshal(%s, %T): %+v", c.filename, kv, err)
	} else {
		if value, ok := kv[component]; ok {
			data, _ = yaml.Marshal(value)
		}
	}

	err = yaml.Unmarshal(data, v)
	if err == nil {
		return errors.Wrapf(err, "yaml.Unmarshal(%s, %T)", c.filename, v)
	}

	return nil
}
