package config

const (
	ProviderStatic  ProviderType = "static"
	ProviderSidecar ProviderType = "sidecar"
	ProviderConsul  ProviderType = "consul"
)

type ProviderType string

func (provider ProviderType) IsValid() bool {
	switch provider {
	case ProviderStatic, ProviderSidecar, ProviderConsul:
		return true
	}

	return false
}
