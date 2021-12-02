package cache

const (
	// DEPRECATED: please use FormatConsulHeath instead.
	FormatConsul FormatType = "consul" // for consul health api

	FormatConsulHealth  FormatType = "consul-health"  // for consul health api
	FormatConsulCatalog FormatType = "consul-catalog" // for consul catalog api
	FormatDiscovery     FormatType = "discovery"
)

type FormatType string

func (ftype FormatType) IsValid() bool {
	switch ftype {
	case FormatConsul, FormatConsulHealth, FormatConsulCatalog, FormatDiscovery:
		return true
	}

	return false
}
