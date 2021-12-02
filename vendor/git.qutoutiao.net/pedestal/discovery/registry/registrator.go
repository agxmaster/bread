package registry

type Registrator interface {
	//注册服务
	Register(*Service, ...RegistratorOption) error
	//注销服务
	Deregister(*Service, ...RegistratorOption) error
}

type RegistratorOption interface {
	IsRegister()
}
