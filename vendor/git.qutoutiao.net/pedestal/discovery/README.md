# discovery 服务注册与发现 SDK.v2

[![pipeline status](https://git.qutoutiao.net/pedestal/discovery/badges/master/pipeline.svg)](https://git.qutoutiao.net/pedestal/discovery/commits/master) [![coverage report](https://git.qutoutiao.net/pedestal/discovery/badges/master/coverage.svg)](https://git.qutoutiao.net/pedestal/discovery/commits/master)


主要用于标准方式部署的 consul，即每个服务机器都需要部署 consul agent 服务！

详细设计见: http://km.qutoutiao.net/pages/viewpage.action?pageId=115085779


## 使用说明

### 服务注册

```go
package main

import (
	"log"
	"sync"

	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

var (
	// 建议在服务初始化时全局初始化一个 registry 对象
	singleRegistry     *discovery.Registry
	singleRegistryOnce sync.Once
)

func main() {
	singleRegistryOnce.Do(func() {
		var err error

		singleRegistry, err = discovery.NewRegistryWithConsul("http://localhost:8500")
		if err != nil {
			panic(err)
		}
	})

	//注册服务
	service, err := singleRegistry.Register(&registry.Service{
		//服务名: 建议ops项目名，不能使用下换线且任何非url safe的字符
		Name: "test-redis",
		//服务注册ip地址
		IP: "10.104.32.79",
		//服务端口
		Port: 9999,
		//tag
		Tags: []string{"test"},
		//元数据
		Meta: map[string]string{
			"my-key": "my-value",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	// 注销服务
	defer service.Deregister()
}
```


### 服务发现

#### 使用 SDK 获取服务节点信息

```go
package main

import (
	"log"
	"sync"

	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/discovery/registry"
)

var (
	// 建议在服务初始化时全局初始化一个 registry 对象
	singleRegistry     *discovery.Registry
	singleRegistryOnce sync.Once
)

func init() {
	singleRegistryOnce.Do(func() {
		var err error

		singleRegistry, err = discovery.NewRegistryWithConsul("http://localhost:8500")
		if err != nil {
			panic(err)
		}
	})
}

func LookupServices() (ipv4s []string, err error) {
	var (
		serviceName = "your-service-name"
		serviceDC   = "your-service-dc"
		serviceTags = []string{"tag1", "tag2"}
	)

	services, err := singleRegistry.LookupServices(serviceName, registry.WithDC(serviceDC), registry.WithTags(serviceTags))
	if err != nil {
		log.Println(err)

		return
	}

	ipv4s = make([]string, len(services))
	for i, service := range services {
		ipv4s[i] = service.Addr()
	}
	log.Println("resolved service", serviceName, ipv4s)

	return
}
```

#### 使用 HTTP 进行服务发现调用

> HTTP 调用推荐使用 [Resty](https://git.qutoutiao.net/pedestal/resty) 封装库！

```go
package main

import (
	"fmt"
    "log"
    "sync"

	"git.qutoutiao.net/pedestal/discovery"
	"git.qutoutiao.net/pedestal/resty"
)

var (
	// 建议全局初始化一个 registry 对象
	singleClient     *http.Client
	singleClientOnce sync.Once
)

func init() {
	singleClientOnce.Do(func() {
		singleClient = resty.New().WithServiceConfig(&config.ServiceConfig{
            Name:    "my-service-name",
            DC:      "dc1",
            Tags:    []string{
                "prd", "v1.0.0",
            },
            Domains: []string{
                "https://www.example.com",
            },
            Connect: &config.ConnectConfig{
                Provider: "consul",
                Addr:     "http://127.0.0.1:8500",
                Enable:   true,
            },
        }).NewHTTPClient()
	})
}

func main() {
    resp, err := singleClient.Get("https://www.example.com/v1/hello")
    if err != nil {
        log.Fatalln(err)
    }
    defer resp.Body.Close()

    // todo
}
```



#### 使用 gRPC 进行服务发现调用

```go
package main

import (
	"fmt"
	"sync"

	"git.qutoutiao.net/pedestal/discovery"
	resolver "git.qutoutiao.net/pedestal/discovery/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

var (
	// 建议全局初始化一个 registry 对象
	singleRegistry     *discovery.Registry
	singleRegistryOnce sync.Once
)

func init() {
	singleRegistryOnce.Do(func() {
		var err error

		singleRegistry, err = discovery.NewRegistryWithConsul("http://localhost:8500")
		if err != nil {
			panic(err)
		}
	})

	// 注册 grpc reslover
	resolver.Register(singleRegistry)
}

func main() {
    serviceName := "my-service-name"
    serviceDC := "dc1"
    serviceTags := []string{
        "prd", "v1.0.0",
    }

	cc, err := resolver.NewDialer(serviceName, serviceDC, serviceTags)
	if err != nil {
		panic(err)
	}

	fmt.Println(cc.GetState())
}
```

