package config

import (
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Loader   *LoaderConfig    `yaml:"loader,omitempty"`
	Connects []*ConnectConfig `yaml:"connects"`
	Services []*ServiceConfig `yaml:"services"`
	Disable  bool             `yaml:"disable"`
}

func NewFromFilename(filename string) (config *Config, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	var tmp struct {
		Config *Config `yaml:"resty"`
	}

	err = yaml.Unmarshal(data, &tmp)
	if err == nil {
		config = tmp.Config
	}

	return
}

type LoaderConfig struct {
	Provider string                 `yaml:"provider"`
	Value    map[string]interface{} `yaml:"value"`
}

func (config *LoaderConfig) IsValid() bool {
	if config == nil {
		return false
	}

	switch config.Provider {
	case "file", "consul", "cc":
		// ignore
	default:
		return false
	}

	return len(config.Value) > 0
}

func (config *LoaderConfig) UnmarshalYAML(v interface{}) error {
	data, err := yaml.Marshal(config.Value)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &v)
}

type ConnectConfig struct {
	Provider ProviderType `yaml:"provider"`
	Addr     string       `yaml:"addr"`
	EDSAddr  string       `yaml:"eds_addr"` // only for consul provider
	Priority int          `yaml:"priority"`
	Enable   bool         `yaml:"enable"`
}

func (config *ConnectConfig) IsValid() bool {
	if config == nil {
		return false
	}

	return config.Enable && config.Provider.IsValid() && len(config.Addr) > 0
}

func (config *ConnectConfig) IsEnabled() bool {
	if config == nil {
		return false
	}

	return config.Enable && len(config.Addr) > 0
}

type ServiceConfig struct {
	Name    string         `yaml:"name"`    // service name registered in consul
	DC      string         `yaml:"dc"`      // for service datacenter
	Tags    []string       `yaml:"tags"`    // for service filter tags
	Port    int            `yaml:"port"`    // overwrite service port if specified, only use by consul connect
	Domains []string       `yaml:"domains"` // for slb domains of service, or ip list
	Connect *ConnectConfig `yaml:"connect"` // overwrite global settings if specified
}

func (config *ServiceConfig) IsConnectValid() bool {
	if config == nil {
		return false
	}

	return config.Connect.IsValid()
}

func (config *ServiceConfig) Match(urlobj *url.URL) bool {
	if config == nil {
		return false
	}

	name := urlobj.Host
	if len(name) == 0 {
		switch {
		case len(urlobj.Scheme) > 0: // for www.example.com:80/api format
			name = urlobj.Scheme

			if len(urlobj.Opaque) > 0 {
				name += ":" + strings.SplitN(urlobj.Opaque, "/", 2)[0]
			}

		case len(urlobj.Path) > 0: // for for www.example.com/api format
			name = strings.SplitN(urlobj.Path, "/", 2)[0]

		}

		if len(name) == 0 {
			return false
		}
	}

	// first, try service name
	if name == config.Name {
		return true
	}

	// avoid service name as domain suffix
	if !strings.ContainsAny(name, ".") {
		return false
	}

	// second, try service domains
	for _, domain := range config.Domains {
		if strings.HasSuffix(domain, name) {
			return true
		}
	}

	return false
}

func (config *ServiceConfig) NormalizeAddr(addr string) string {
	if config.Port <= 0 {
		return addr
	}

	ip2port := strings.SplitN(addr, ":", 2)
	switch len(ip2port) {
	case 1:
		ip2port = append(ip2port, strconv.Itoa(config.Port))
	case 2:
		ip2port[1] = strconv.Itoa(config.Port)
	}

	return strings.Join(ip2port, ":")
}
