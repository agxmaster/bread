package discovery

const (
	DefaultTempDir = "discovery-local"
)

const (
	FailBack FailType = 0
	FailFast FailType = 1
)

type FailType int

func (ft FailType) IsFailFast() bool {
	return ft == FailFast
}

const (
	RegistryConsul RegistryType = "consul"
	RegistryFile   RegistryType = "file"
)

type RegistryType string

func (rtype RegistryType) IsValid() bool {
	switch rtype {
	case RegistryConsul, RegistryFile:
		return true
	}

	return false
}

const (
	DefaultSentinelAddr = "http://infra-techcenter-arch-sentinel-prd.5qtt.cn"
)
