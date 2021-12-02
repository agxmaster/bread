package registry

import (
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

func convertName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

// toRegistryName qms以前注册的服务名
func toRegistryName(name string, proto protocol.Protocol) string {
	if proto == protocol.ProtocHTTP {
		return convertName(name)
	}
	return convertName(name) + "-" + proto.String()
}

// toNewRegistryName 最新规范的服务名
func toStdRegistryName(name string, proto protocol.Protocol, env qenv.Env) string {
	name = convertName(name)
	if proto != protocol.ProtocHTTP {
		name += "-" + proto.String() + "-proto"
	}
	if env != qenv.PRD && env != qenv.QA {
		name += "-" + env.String() + "-env"
	}
	return name
}
