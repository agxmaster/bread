package iface

type OldFields interface {
	SetFields(map[string]interface{})
	GetFields() map[string]interface{}
}
