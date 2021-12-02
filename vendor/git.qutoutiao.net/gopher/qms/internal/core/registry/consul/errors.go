package consul

import "fmt"

var (
	ErrNewRegistry = fmt.Errorf("new registry error")
	ErrDeregister  = fmt.Errorf("deregister error")
	ErrRegister    = fmt.Errorf("register service error")
)
