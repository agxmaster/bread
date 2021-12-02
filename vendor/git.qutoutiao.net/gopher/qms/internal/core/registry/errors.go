package registry

import "fmt"

var (
	ErrNotFoundBuilder = fmt.Errorf("not found builder")
	ErrNotFoundService = fmt.Errorf("not found services")
	ErrGetConfig       = fmt.Errorf("can't get config")
)
