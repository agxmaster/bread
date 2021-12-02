package transport

import "net/http"

const (
	ProviderStatic  ProviderType = "static"
	ProviderSRV     ProviderType = "srv"
	ProviderSidecar ProviderType = "sidecar"
	ProviderConsul  ProviderType = "consul"
)

type ProviderType string

func (provider ProviderType) IsValid() bool {
	switch provider {
	case ProviderStatic, ProviderSRV, ProviderSidecar, ProviderConsul:
		return true
	}

	return false
}

type Interface interface {
	http.RoundTripper

	IsValid() bool
	Check()
}
