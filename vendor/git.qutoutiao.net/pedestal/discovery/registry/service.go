package registry

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"git.qutoutiao.net/pedestal/discovery/errors"
	"git.qutoutiao.net/pedestal/discovery/logger"
	"github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/go-sockaddr/template"
)

const (
	DefaultHostname      = "default"
	DefaultServiceWeight = 100
)

// Service
type Service struct {
	ID         string            `discovery:"可选,服务id"`
	Name       string            `discovery:"必填,服务名"`
	IP         string            `discovery:"可选,默认拿en0的地址"`
	IPTemplate string            `discovery:"可选,可以用于指定特殊网卡" json:"IPTemplate,omitempty"`
	Tags       []string          `discovery:"可选,标签"`
	Port       int               `discovery:"必填,端口"`
	Weight     int32             `discovery:"可选,权重"`
	Meta       map[string]string `discovery:"可选,自定义元数据"`

	once sync.Once
}

func (s *Service) FillWithDefaults() {
	s.once.Do(func() {
		// adjust service ip if empty
		if len(s.IP) == 0 {
			// try ip template if provided
			if len(s.IPTemplate) > 0 {
				ipv4, err := template.Parse(s.IPTemplate)
				if err == nil {
					s.IP = ipv4
				} else {
					logger.Errorf("template.Parse(%s): %v", s.IPTemplate, err)
				}
			}

			if len(s.IP) == 0 {
				ipv4, err := sockaddr.GetPrivateIP()
				if err == nil {
					s.IP = ipv4
				} else {
					logger.Errorf("sockaddr.GetPrivateIP(): %v", err)
				}
			}
		}

		// adjust service id
		if len(s.ID) == 0 {
			hostname, err := os.Hostname()
			if err != nil {
				hostname = strconv.FormatInt(int64(s.Port), 10)
			}

			s.ID = s.Name + "~" + s.IP + "~" + hostname
		}

		// avoid panic with write to nil map
		if s.Meta == nil {
			s.Meta = make(map[string]string)
		}

		if s.Weight > 0 {
			s.Meta["weight"] = strconv.FormatUint(uint64(s.Weight), 10)
		}
	})
}

func (s *Service) Valid() error {
	s.FillWithDefaults()

	if len(s.Name) == 0 {
		return errors.ErrInvalidService
	}

	name := s.ServiceName()
	if s.Name != name {
		return fmt.Errorf("invalid format of service name, it should be %s", name)
	}

	return nil
}

// Addr returns ip:port of service.
func (s *Service) Addr() string {
	s.FillWithDefaults()

	return s.IP + ":" + strconv.FormatInt(int64(s.Port), 10)
}

// ServiceID returns id of service.
func (s *Service) ServiceID() string {
	s.FillWithDefaults()

	return s.ID
}

// ServiceName returns normalized name of service.
func (s *Service) ServiceName() string {
	var names []string
	for _, name := range strings.Split(s.Name, "_") {
		names = append(names, strings.Trim(name, "-"))
	}

	return strings.Join(names, "-")
}

func (s *Service) ServiceIP() string {
	s.FillWithDefaults()

	return s.IP
}

func (s *Service) ServiceWeight() int {
	if s.Weight > 0 {
		return int(s.Weight)
	}

	if s.Meta != nil {
		if weight, ok := s.Meta["weight"]; ok {
			i32, err := strconv.ParseInt(weight, 10, 32)
			if err == nil {
				return int(i32)
			}
		}
	}

	return DefaultServiceWeight
}
