package consul

import (
	"strconv"

	"github.com/hashicorp/consul/api"

	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

// IsConsulHealthMaint returns true if service check status is api.HealthMaint or
// key's meta contains status=offline.
//
// NOTE: It does not check node's metadata of status=offline setting.
func IsConsulHealthMaint(entry *api.ServiceEntry) bool {
	switch entry.Checks.AggregatedStatus() {
	case api.HealthMaint:
		return true
	}

	if entry.Service == nil || len(entry.Service.Meta) == 0 {
		return false
	}

	// for custom maint with meta
	return entry.Service.Meta["status"] == "offline"
}

// IsConsulCatalogMaint returns true if service check status is api.HealthMaint or
// key's meta contains status=offline.
//
// NOTE: It does not check node's metadata of status=offline setting.
func IsConsulCatalogMaint(entry *api.CatalogService) bool {
	switch entry.Checks.AggregatedStatus() {
	case api.HealthMaint:
		return true
	}

	if len(entry.ServiceMeta) == 0 {
		return false
	}

	// for custom maint with meta
	return entry.ServiceMeta["status"] == "offline"
}

// ParseConsulHealth converts *api.ServiceEntry resolved from health api to *registry.Service.
func ParseConsulHealth(entries []*api.ServiceEntry) []*registry.Service {
	services := make([]*registry.Service, 0, len(entries))

	for _, entry := range entries {
		if entry.Service == nil {
			continue
		}

		// avoid race by copying meta
		meta := make(map[string]string)
		for k, v := range entry.Service.Meta {
			meta[k] = v
		}

		// calc weight of service
		weight := int32(registry.DefaultServiceWeight)
		if entry.Service.Weights.Passing > 0 {
			weight = int32(entry.Service.Weights.Passing)
		}
		if metaWeight, ok := meta["weight"]; ok {
			i32, err := strconv.ParseInt(metaWeight, 10, 32)
			if err != nil {
				logger.Errorf("cannot parse weight(%v) of service(%v): %v", weight, entry.Service.Service, err)
			} else {
				weight = int32(i32)
			}
		}

		services = append(services, &registry.Service{
			ID:     entry.Service.ID,
			Name:   entry.Service.Service,
			IP:     entry.Service.Address,
			Port:   entry.Service.Port,
			Tags:   entry.Service.Tags[:],
			Meta:   meta,
			Weight: weight,
		})
	}

	return services
}

// ParseConsulCatalog converts *api.CatalogService resolved from catalog api to *registry.Service.
func ParseConsulCatalog(catalogServices []*api.CatalogService) []*registry.Service {
	services := make([]*registry.Service, 0, len(catalogServices))

	for _, service := range catalogServices {
		// avoid race by copying meta
		meta := make(map[string]string)
		for k, v := range service.ServiceMeta {
			meta[k] = v
		}

		weight := int32(registry.DefaultServiceWeight)
		if service.ServiceWeights.Passing > 0 {
			weight = int32(service.ServiceWeights.Passing)
		}
		if metaWeight, ok := meta["weight"]; ok {
			i32, err := strconv.ParseInt(metaWeight, 10, 32)
			if err != nil {
				logger.Errorf("cannot parse weight(%v) of service(%s): %v", weight, service.ServiceID, err)
			} else {
				weight = int32(i32)
			}
		}

		services = append(services, &registry.Service{
			ID:     service.ServiceID,
			Name:   service.ServiceName,
			IP:     service.ServiceAddress,
			Port:   service.ServicePort,
			Tags:   service.ServiceTags[:],
			Meta:   meta,
			Weight: weight,
		})
	}

	return services
}

func ReduceConsulHealthWithPassingOnly(entries []*api.ServiceEntry) []*api.ServiceEntry {
	entries = ReduceConsulHealthWithoutMaint(entries)

	passingEntries := make([]*api.ServiceEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Service == nil {
			continue
		}

		switch entry.Checks.AggregatedStatus() {
		case api.HealthPassing, api.HealthWarning:
			passingEntries = append(passingEntries, entry)
		}
	}

	return passingEntries
}

func ReduceConsulHealthWithoutMaint(entries []*api.ServiceEntry) []*api.ServiceEntry {
	reduceMap := make(map[string]int)
	newEntries := make([]*api.ServiceEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Service == nil {
			continue
		}

		// filter out entry with api.HealthMaint status
		if IsConsulHealthMaint(entry) {
			continue
		}

		if idx, ok := reduceMap[entry.Service.ID]; ok {
			if newEntries[idx].Service.ModifyIndex < entry.Service.ModifyIndex {
				newEntries[idx] = entry
			}
			continue
		}

		newEntries = append(newEntries, entry)
		reduceMap[entry.Service.ID] = len(newEntries) - 1
	}

	return newEntries
}

func ReduceConsulHealthWithMaint(entries []*api.ServiceEntry) []*api.ServiceEntry {
	reduceMap := make(map[string]int)
	newEntries := make([]*api.ServiceEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Service == nil {
			continue
		}

		// filter out entry without api.HealthMaint status
		if !IsConsulHealthMaint(entry) {
			continue
		}

		if idx, ok := reduceMap[entry.Service.ID]; ok {
			if newEntries[idx].Service.ModifyIndex < entry.Service.ModifyIndex {
				newEntries[idx] = entry
			}
			continue
		}

		newEntries = append(newEntries, entry)
		reduceMap[entry.Service.ID] = len(newEntries) - 1
	}

	return newEntries
}

func ReduceConsulCatalogWithoutMaint(services []*api.CatalogService, passingOnly bool) []*api.CatalogService {
	reduceMap := make(map[string]int)
	newServices := make([]*api.CatalogService, 0, len(services))
	for _, service := range services {
		status := service.Checks.AggregatedStatus()
		if passingOnly && (status != api.HealthPassing) && (status != api.HealthWarning) {
			continue
		}

		// filter out key with api.HealthMaint status
		if IsConsulCatalogMaint(service) {
			continue
		}

		if idx, ok := reduceMap[service.ServiceID]; ok {
			if newServices[idx].ModifyIndex < service.ModifyIndex {
				newServices[idx] = service
			}
			continue
		}

		newServices = append(newServices, service)
		reduceMap[service.ServiceID] = len(newServices) - 1
	}

	return newServices
}

func ReduceConsulCatalogWithMaint(services []*api.CatalogService, passingOnly bool) []*api.CatalogService {
	reduceMap := make(map[string]int)
	newServices := make([]*api.CatalogService, 0, len(services))
	for _, service := range services {
		status := service.Checks.AggregatedStatus()
		if passingOnly && (status != api.HealthPassing) && (status != api.HealthWarning) {
			continue
		}

		// filter out key without api.HealthMaint status
		if !IsConsulCatalogMaint(service) {
			continue
		}

		if idx, ok := reduceMap[service.ServiceID]; ok {
			if newServices[idx].ModifyIndex < service.ModifyIndex {
				newServices[idx] = service
			}
			continue
		}

		newServices = append(newServices, service)
		reduceMap[service.ServiceID] = len(newServices) - 1
	}

	return newServices
}
