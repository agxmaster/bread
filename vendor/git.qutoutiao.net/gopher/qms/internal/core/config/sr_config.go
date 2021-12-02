package config

import (
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// constant for service registry parameters
const (
	DefaultSRAddressQA  = "http://registry-qa.qutoutiao.net"
	DefaultSRAddressPRD = "http://127.0.0.1:8500"
)

// GetRegistratorType returns the Type of service registry
func GetRegistratorType() string {
	if GlobalDefinition.Qms.Service.Registry.Registrator.Type != "" {
		return GlobalDefinition.Qms.Service.Registry.Registrator.Type
	}
	return GlobalDefinition.Qms.Service.Registry.Type
}

// GetRegistratorAddress returns the Address of service registry
func GetRegistratorAddress() string {
	if GlobalDefinition.Qms.Service.Registry.Registrator.Address != "" {
		return GlobalDefinition.Qms.Service.Registry.Registrator.Address
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

// GetRegistratorScope returns the Scope of service registry
func GetRegistratorScope() string {
	if GlobalDefinition.Qms.Service.Registry.Registrator.Scope == "" {
		GlobalDefinition.Qms.Service.Registry.Registrator.Scope = common.ScopeFull
	}
	return GlobalDefinition.Qms.Service.Registry.Scope
}

// GetRegistratorAutoRegister returns if auto register service
func GetRegistratorAutoRegister() bool {
	if GlobalDefinition.Qms.Service.Registry.DisableRegister {
		return false
	}
	if GlobalDefinition.Qms.Service.Registry.Registrator.AutoRegister == "manual" {
		return false
	}
	if GlobalDefinition.Qms.Service.Registry.AutoRegister == "manual" {
		return false
	}
	return true
}

// GetRegistratorTenant returns the Tenant of service registry
func GetRegistratorTenant() string {
	if GlobalDefinition.Qms.Service.Registry.Registrator.Tenant != "" {
		return GlobalDefinition.Qms.Service.Registry.Registrator.Tenant
	}
	return GlobalDefinition.Qms.Service.Registry.Tenant
}

// GetRegistratorAPIVersion returns the APIVersion of service registry
func GetRegistratorAPIVersion() string {
	if GlobalDefinition.Qms.Service.Registry.Registrator.APIVersion.Version != "" {
		return GlobalDefinition.Qms.Service.Registry.Registrator.APIVersion.Version
	}
	return GlobalDefinition.Qms.Service.Registry.APIVersion.Version
}
