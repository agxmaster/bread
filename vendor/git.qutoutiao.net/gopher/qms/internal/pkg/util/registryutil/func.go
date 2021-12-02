package registryutil

import (
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

func convertName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

// toNewRegistryName 最新规范的服务名
func ToStdRegistryName(name string, proto protocol.Protocol, env qenv.Env) string {
	name = convertName(name)
	if proto != protocol.ProtocHTTP {
		name += "-" + proto.String() + "-proto"
	}
	if env != qenv.PRD && env != qenv.QA {
		name += "-" + env.String() + "-env"
	}
	return name
}
