package config

// GetContractDiscoveryType returns the Type of contract discovery registry
func GetContractDiscoveryType() string {
	if GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Type != "" {
		return GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Type
	}
	return GlobalDefinition.Qms.Service.Registry.Type
}

// GetContractDiscoveryAddress returns the Address of contract discovery registry
func GetContractDiscoveryAddress() string {
	if GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Address != "" {
		return GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Address
	}
	return GlobalDefinition.Qms.Service.Registry.Address
}

// GetContractDiscoveryTenant returns the Tenant of contract discovery registry
func GetContractDiscoveryTenant() string {
	if GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Tenant != "" {
		return GlobalDefinition.Qms.Service.Registry.ContractDiscovery.Tenant
	}
	return GlobalDefinition.Qms.Service.Registry.Tenant
}

// GetContractDiscoveryAPIVersion returns the APIVersion of contract discovery registry
func GetContractDiscoveryAPIVersion() string {
	if GlobalDefinition.Qms.Service.Registry.ContractDiscovery.APIVersion.Version != "" {
		return GlobalDefinition.Qms.Service.Registry.ContractDiscovery.APIVersion.Version
	}
	return GlobalDefinition.Qms.Service.Registry.APIVersion.Version
}
