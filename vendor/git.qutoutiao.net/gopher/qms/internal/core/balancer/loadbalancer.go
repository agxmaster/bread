package balancer

// PickerBuilder creates balancer.Picker.
type LoadbalancerBuilder interface {
	Build() Loadbalancer
	Name() string
}

// 兼容rebalancer定义的Loadbalancer接口
type Loadbalancer interface {
	Upsert(id string, weight int, node interface{}) error
	Next() (interface{}, error)
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
