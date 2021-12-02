package registry

import (
	"sort"
	"strings"

	"git.qutoutiao.net/pedestal/discovery/errors"
)

// ServiceKey defines service with query meta
type ServiceKey struct {
	Name string
	Tags string
	DC   string
}

func NewServiceKey(name string, tags []string, dc string) ServiceKey {
	sort.Strings(tags)

	tag := strings.Join(tags, ":")

	return ServiceKey{
		Name: name,
		Tags: tag,
		DC:   dc,
	}
}

func (key *ServiceKey) ToString() string {
	fields := make([]string, 0)
	if len(key.Tags) > 0 {
		fields = append(fields, key.Tags)
	}

	fields = append(fields, key.Name)
	fields = append(fields, "service")

	if len(key.DC) > 0 {
		fields = append(fields, key.DC)
	}

	return strings.Join(fields, ".")

}

// key formatted in [<tags>.]<service name>.service.[.<consul datacenter>]
func ParseServiceKey(key string) (*ServiceKey, error) {
	fields := strings.Split(key, ".")

	findService := func(fields []string) int {
		for i, field := range fields {
			if field == "service" {
				return i
			}
		}
		return -1
	}

	idx := findService(fields)
	if idx <= 1 {
		return nil, errors.Wrap(errors.ErrArgument)
	}

	serviceKey := &ServiceKey{}

	serviceKey.Name = fields[idx-1]
	if idx-1 > 0 {
		serviceKey.Tags = strings.Join(fields[:idx-1], ".")
	}
	if idx+1 > len(fields) {
		serviceKey.DC = strings.Join(fields[idx+1:], ".")
	}

	return serviceKey, nil
}
