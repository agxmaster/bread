package registry

type RegisterOpt func(*CommonRegistratorOption)

func (RegisterOpt) IsRegister() {
}

type CommonRegistratorOption struct {
	Checks   []*HealthCheck
	Metadata map[string]string
}

func WithHealthCheck(check *HealthCheck) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		o.Checks = append(o.Checks, check)
	}
}

func WithTCPHealthCheck(check *TCPHealthCheck) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		o.Checks = append(o.Checks, &HealthCheck{
			Type:     HealthTypeTCP,
			Name:     check.Name,
			URI:      check.Addr,
			Interval: check.Interval,
			Status:   check.Status,
		})
	}
}

func WithHTTPHealthCheck(check *HTTPHealthCheck) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		o.Checks = append(o.Checks, &HealthCheck{
			Type:     HealthTypeHTTP,
			Name:     check.Name,
			URI:      check.URI,
			Method:   check.Method,
			Header:   check.Header,
			Interval: check.Interval,
			Status:   check.Status,
		})
	}
}

// 注册时，特殊的metadata
// 机器所在云环境 目前为aliyun跟tencent
func WithCloud(cloud string) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["cloud"] = cloud
	}
}

//机器所在分区，默认通过hostname获取
func WithZone(zone string) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["zone"] = zone
	}
}

//k8s或者vm环境, 默认为vm
func WithContainer(container string) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["container"] = container
	}
}

func WithRegistry(registry string) RegisterOpt {
	return func(o *CommonRegistratorOption) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]string)
		}
		o.Metadata["registry"] = registry
	}
}

//=====================discovery=====================
type CommonDiscoveryOption struct {
	DC   string
	Tags []string
}

type DiscoveryOpt func(*CommonDiscoveryOption)

func (DiscoveryOpt) IsDiscovery() {
}

func WithDC(dc string) DiscoveryOpt {
	return func(o *CommonDiscoveryOption) {
		o.DC = dc
	}
}

func WithTags(tags []string) DiscoveryOpt {
	return func(o *CommonDiscoveryOption) {
		o.Tags = tags
	}
}

func NewCommonDiscoveryOption(opts ...DiscoveryOption) *CommonDiscoveryOption {
	o := new(CommonDiscoveryOption)
	for _, opt := range opts {
		switch opt := opt.(type) {
		case DiscoveryOpt:
			opt(o)
		}
	}
	return o
}
