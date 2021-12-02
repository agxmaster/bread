package sentinel

import "github.com/hashicorp/consul/api"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type SnapshotResponse struct {
	UseSentinel bool                  `json:"use_sentinel"`
	Data        []*api.CatalogService `json:"data"`
}
