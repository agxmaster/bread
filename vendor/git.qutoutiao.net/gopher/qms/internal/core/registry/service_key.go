package registry

import (
	"sort"
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// ServiceKey 通用的ServiceKey接口实现
type ServiceKey struct {
	name            string // 服务发现使用的名称
	originName      string
	registryName    string
	stdRegistryName string
	dc              string
	tags            []string
	proto           protocol.Protocol
	env             qenv.Env
}

// NewServiceKey 必须指定服务名[契约式设计 对name不做校验]
func NewServiceKey(name, dc string, tags []string, proto protocol.Protocol, env qenv.Env) *ServiceKey {
	key := &ServiceKey{
		originName: name,
		dc:         dc,
		tags:       tags,
		proto:      proto,
		env:        env,
	}
	key.registryName = toRegistryName(name, proto)
	key.stdRegistryName = toStdRegistryName(name, proto, env)

	sort.Strings(key.tags)
	return key
}

func (sk *ServiceKey) Name() string {
	return sk.name
}

func (sk *ServiceKey) DC() string {
	return sk.dc
}

// 需要判断tags是否为nil
func (sk *ServiceKey) Tags() []string {
	return sk.tags
}

// service:rest-dc:qa-tags:qms:v1.3.5,
func (sk *ServiceKey) String() string {
	tag := strings.Join(sk.tags, ",")
	return strings.Join([]string{"service:" + sk.originName, "dc:" + sk.dc, "tags:" + tag}, "-")
}
