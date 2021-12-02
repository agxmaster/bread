package registryutil

import "git.qutoutiao.net/gopher/qms/internal/pkg/runtime"

func GetContainer() string {
	if runtime.InsideDocker {
		return "k8s"
	}
	return "vm"
}
