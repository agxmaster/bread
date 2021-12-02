package config

import (
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// GetServiceDiscoveryType returns the Type of SD registry
func GetServiceDiscoveryType() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Type != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Type
	}
	return GlobalDefinition.Qms.Service.Registry.Type
}

// GetServiceDiscoveryAddress returns the Address of SD registry
func GetServiceDiscoveryAddress() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Address != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Address
	}
	if GlobalDefinition.Qms.Service.Registry.Address == "" {
		e := qenv.Get()
		if e == qenv.PRD || e == qenv.PRE {
			GlobalDefinition.Qms.Service.Registry.Address = DefaultSRAddressPRD
		} else {
			GlobalDefinition.Qms.Service.Registry.Address = DefaultSRAddressQA
		}
	}
	return GlobalDefinition.Qms.Service.Registry.Address
}

// GetServiceDiscoveryRefreshInterval returns the RefreshInterval of SD registry
func GetServiceDiscoveryRefreshInterval() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.RefreshInterval != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.RefreshInterval
	}
	return GlobalDefinition.Qms.Service.Registry.RefreshInterval
}

// GetServiceDiscoveryWatch returns the Watch of SD registry
func GetServiceDiscoveryWatch() bool {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Watch {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Watch
	}
	return GlobalDefinition.Qms.Service.Registry.Watch
}

// GetServiceDiscoveryTenant returns the Tenant of SD registry
func GetServiceDiscoveryTenant() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Tenant != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.Tenant
	}
	return GlobalDefinition.Qms.Service.Registry.Tenant
}

// GetServiceDiscoveryAPIVersion returns the APIVersion of SD registry
func GetServiceDiscoveryAPIVersion() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.APIVersion.Version != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.APIVersion.Version
	}
	return GlobalDefinition.Qms.Service.Registry.APIVersion.Version
}

// GetServiceDiscoveryDisable returns the Disable of SD registry
func GetServiceDiscoveryDisable() bool {
	return qconf.GetBool("qms.service.registry.serviceDiscovery.disabled", false)
}

// GetServiceDiscoveryHealthCheck returns the HealthCheck of SD registry
func GetServiceDiscoveryHealthCheck() bool {
	if b := qconf.GetBool("qms.service.registry.serviceDiscovery.healthCheck", false); b {
		return b
	}
	return qconf.GetBool("qms.service.registry.healthCheck", false)
}

// DefaultConfigPath set the default config path
const DefaultConfigPath = "/etc/.kube/config"

// GetServiceDiscoveryConfigPath returns the configpath of SD registry
func GetServiceDiscoveryConfigPath() string {
	if GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.ConfigPath != "" {
		return GlobalDefinition.Qms.Service.Registry.ServiceDiscovery.ConfigPath
	}
	return DefaultConfigPath
}
