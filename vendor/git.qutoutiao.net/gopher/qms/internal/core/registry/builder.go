package registry

import (
	"sync"
)

type Builder interface {
	Build(config *Config) (Registry, error)
}

var builders sync.Map // map[name]Builder

func RegisterBuilder(name string, builder Builder) {
	builders.Store(name, newRegistryBuilder(builder))
}

func getBuilder(name string) Builder {
	if b, ok := builders.Load(name); ok {
		return b.(Builder)
	}
	return nil
}
