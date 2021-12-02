package balancer

import (
	"fmt"
	"sync"
)

var (
	balancers = make(map[string]Balancer)
	sl        sync.RWMutex
)

func newBalancer(name string, r Resolver) (Balancer, error) {
	builder := getBuilder(name)
	if builder == nil {
		return nil, fmt.Errorf("don't have balance builder plugin %s", name)
	}

	return builder.Build(r), nil
}

func GetBalancer(name string) (b Balancer, err error) {
	sl.RLock()
	b, ok := balancers[name]
	sl.RUnlock()
	if !ok {
		b, err = newBalancer(name, getResolver())
		if err != nil {
			return nil, err
		}

		sl.Lock()
		balancers[name] = b
		sl.Unlock()
	}
	return b, nil
}

func Enable() {
	//config.GetLoadBalancing().Enabled = true 现在默认是开启的
}

//for k, v := range instance.EndpointsMap {
//	serviceName := serviceName
//	if k != "rest" {
//		serviceName += "-" + k
//	}
//
//	ipPort := strings.Split(v, ":")
//	if len(ipPort) != 2 {
//		return "", errors.Errorf("failed to get port %#v ", instance.EndpointsMap)
//	}
//
//	//使用传入的endpoint,假如没有传入,就用获取的ip
//	var ip string
//	if len(ipPort[0]) != 0 {
//		ip = ipPort[0]
//	} else {
//		ip = iputil.GetLocalIP()
//	}
//	instanceID = ip
//
//	port, err := strconv.Atoi(ipPort[1])
//	if err != err {
//		return "", errors.Errorf("recover port failed %s ", ipPort[1])
//	}
//
//	endpoint := ip + ":" + strconv.FormatInt(int64(port), 10)
//	hc := registryutil.NewHealthCheck(service.AppID, endpoint)
//	tags := []string{
//		k,
//		fmt.Sprintf("%s:%s", r.microService.Framework.Name, r.microService.Framework.Version),
//		service.Env,
//		service.Version,
//	}
//	//注册服务
//	qlog.Infof("register service(name=%s) to consul(%s)", serviceName, r.consulAddr)
//	deRegistorService, err := r.r.Register(&quregistry.Service{
//		//服务名: 建议ops项目名，不能使用下换线且任何非url safe的字符
//		Name: serviceName,
//		//服务注册ip地址
//		IP: ip,
//		//服务端口
//		Port:   port,
//		Weight: registryutil.DefaultWeight,
//		Tags:   tags,
//		Meta: map[string]string{
//			"protoc":                 k,
//			registryutil.Region:      "",
//			registryutil.Zone:        "",
//			registryutil.Container:   container,
//			registryutil.Healthcheck: hc.String(),
//			registryutil.Status:      registryutil.Online.String(),
//			registryutil.Weight:      "100",
//			registryutil.Regtime:     strconv.FormatInt(time.Now().Unix(), 10),
//		},
//	}, quregistry.WithHTTPHealthCheck(&quregistry.HTTPHealthCheck{
//		Name:     registryutil.Healthcheck,
//		URI:      hc.HTTP,
//		Interval: registryutil.DefaultHCInterval,
//		//Status:   quregistry.HealthPassing,
//		Method: hc.Method,
//		Header: hc.Header,
//	}))
//
//	//annotation: 注册多个，有失败就返回
//	if err != nil {
//		qlog.Errorf("registor failed v: err:%s", err.Error())
//		return "", err
//	}
//
//	r.deRegistors[k] = deRegistorService
//}

//return instanceID, nil
