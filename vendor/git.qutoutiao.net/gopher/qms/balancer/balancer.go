package balancer

import (
	"context"
)

// PickerBuilder creates balancer.Picker.
type PickerBuilder interface {
	Build(services []*Service) Picker
	Name() string
}

type Picker interface {
	Pick(ctx context.Context, opts *PickOptions) (*Service, error)
}

type PickOptions struct {
	SourceService string      // 本服务
	RemoteService string      // 远程调用服务
	Protocol      string      // rest or grpc
	Path          string      // [rest: url-path] [grpc: FullMethod]
	Args          interface{} // [rest: http.Request] [grpc: input]
	Datacenter    string      // 指定DC
	Tags          []string    // 指定Tag
}

type Service struct {
	ID       string            // DiscoveryID
	Name     string            // 服务名称
	IP       string            // IP[必填]
	Port     int               // 端口[必填]
	Endpoint string            // IP:Port
	Weight   int32             // 权重
	Tags     []string          // Tags
	Meta     map[string]string // Meta data
}
