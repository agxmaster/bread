package outputs

import "github.com/go-bread/iface/entity_query"

type OutputField struct {
	TableField string
	Table      string
	OutPut     string
	F          entity_query.CallbackFunc
}

type Callbacks map[string]entity_query.CallbackFunc
