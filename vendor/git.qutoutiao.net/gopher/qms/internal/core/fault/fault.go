package fault

import (
	"git.qutoutiao.net/gopher/qms/internal/core/config/model"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
)

// InjectFault inject fault
type InjectFault func(model.Fault, *invocation.Invocation) error

// Injectors fault injectors
var Injectors = make(map[string]InjectFault)

//Fault fault injection error
type Fault struct {
	Message string
}

func (e Fault) Error() string {
	return e.Message
}

// InstallFaultInjectionPlugin install fault injection plugin
func InstallFaultInjectionPlugin(name string, f InjectFault) {
	Injectors[name] = f
}

func init() {
	InstallFaultInjectionPlugin("rest", faultInject)
	InstallFaultInjectionPlugin("highway", faultInject)
	InstallFaultInjectionPlugin("dubbo", faultInject)
}

func faultInject(rule model.Fault, inv *invocation.Invocation) error {
	return ValidateAndApplyFault(&rule, inv)
}
