package actionutil

type Action int

const (
	ActionUnknown Action = iota
	ActionReload
	ActionClose
)
