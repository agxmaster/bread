## v1.3.20 2021-01-22
- Fixed:
    - grpc 服务通过ctx.Logger.XXX()打印requestID
    - grpc 支持mesh

## v1.3.19 2020-12-25
- Changed
    - 使用最新的ReBalancer，增加节点熔断保护；
    - proto field option增加jsontag；
    - 服务关闭时可以指定容器等待时间；
    - trace tag过长被截取，可以指定buf size；
    - 增加mesh_service配置，可以指定mesh-header，保证路由准确；
    
- Fixed:
    - health_check误报，由于共用header，导致ping判断逻辑有问题；
    - read body failed: context canceled bug；

## v1.3.18 2020-11-04
- Changed
    - (依赖最新的gormv1.0.8)[https://git.qutoutiao.net/golib/gorm]
    
- Fixed:
    - qms-gen生成代码模板使用最新的SDK编写。

## v1.3.17 2020-11-04
- Changed
    - [统一SDK(gorm、redis、resty)](https://git.qutoutiao.net/golib)
    - 删除retry
    - 支持sidecar自定义address，方便dev调试
    
- Fixed:
    - 修改loadbalancer日志级别为debug

## v1.3.15 2020-10-21
- Changed
    - recovery ctx.output panic
    - gin server添加apidoc和pprof 方便开发使用
    - 关闭上传swaggerAPI到yapi平台
    
- Fixed:
    - 容器忽略自动上报metrics，容器会自动收集
    - 增加service mesh名称打印

## v1.3.14 2020-09-07
- Changed
    - [qms支持服务注册发现规范](https://km.qutoutiao.net/pages/viewpage.action?pageId=88485252)
    - [上线rebalancer新负载均衡器](https://git.qutoutiao.net/gopher/rebalancer)
    - metrics gauge支持Reset接口
    
- Fixed:
    - 获取不存在的redis client导致goroutine增加

- 注意：
  - grpc请求上游服务，只需要传递PaasID即可，不需要添加-grpc后缀(向前兼容)
  - 老服务支持多个rest端口、或多个grpc端口，则谨慎升级，请提前联系。
        
## v1.3.13 2020-08-06
- Changed
    - [增加native HTTP服务,管理metrics、pprof、apidoc](https://git.qutoutiao.net/gopher/qms/blob/master/docs/development.md#412-%E4%BB%80%E4%B9%88%E6%98%AFnative%E6%9C%8D%E5%8A%A1%E5%9C%BA%E6%99%AF%E6%98%AF)
    - 使用promhttp.Handler ，便于其他SDK的metrics可以正确收集
    - rest请求报错, 增加完整URL的日志, 或通过resp.Request.URL获取完成URL
    - 引入pkg/json库，默认使用[高性能json库](https://github.com/json-iterator/go)
    
- Fixed:
    - HTTP请求不匹配的路由全部聚合到/404
    - [使用lru优化httpcache，避免内存被打爆](https://git.qutoutiao.net/gopher/qms/blob/master/docs/std-confs/upstream.yaml#L14)
    - graceful时不再reset qconf内存
    - pilot配置问题导致服务不能启动，默认开启pilot配置。
    
native端口默认是9080对应容器工单的默认监控端口。

## v1.3.12 2020-07-20
- Fixed:
    - hotfix of pilot block with wrong config

## v1.3.11 2020-07-10
- Changed
    - 服务发现支持Pilot
    - [http.Server添加超时](https://git.qutoutiao.net/gopher/qms/commit/7eb83a0ee2290929a04843eb04a2de6759c0ee70)
    - [overseer模式，服务异常退出增加错误日志](https://git.qutoutiao.net/gopher/qms/commit/cd46a8b08d59a692eb363ce673f7ada318377e35)
	- [增加command示例](https://git.qutoutiao.net/gopher/qms/tree/master/examples/command)
	 
- Fixed:
	- WithConfigDir失效问题
	- protobuf升级导致json输出由下划线变成驼峰
	- [qms-gen调整配置文件的生成](https://git.qutoutiao.net/gopher/qms/commit/9e936acf061abfe9474cecf877caec0c0bb021e5)

## v1.3.10 2020-06-09
- Changed
	- 实现debug/pprof/block
	- 实现debug/pprof/mutex
	- 更新discovery为v1.13.4
	 
- Fixed:
	- autometrcis注册使用查询的url，导致不会注册。
	- PassID乱码 
	

## v1.3.9 2020-05-26
- Changed
	- 梳理和收敛[配置](https://git.qutoutiao.net/gopher/qms/tree/master/docs/std-confs)。
	- 服务维度管理服务治理功能，增加禁用某个服务服务注册的功能。
	- 移除框架和配置强耦合，为SDK做准备(app.yaml非必须)。
	- 规范配置，框架配置采用qms前缀，key采用下划线，且key不区分大小写。
	- 配置管理底层采用viper。
	- accesslog增加query，并修改duration为duration_ms字段，便于操作SLS、降低索引成本。
	 
- Fixed:
	- 移除archaius(华为)配置管理，避免archaius升级后不向后兼容问题。

## v1.3.8 2020-05-09
- Fixed
	- 开启熔断且并发数较多时会导致[Max Concurrency]的错误
	- 修改配置version不生效的问题

## v1.3.7	2020-04-29
- Fixed
	- go mod tidy命令导致报错

## v1.3.6	2020-04-28
- Changed
	- 服务发现支持指定dc和tags。
	- 增加[upstream.yaml](https://git.qutoutiao.net/gopher/qms/blob/master/docs/std-confs/upstream.yaml)放置访问上游服务的服务治理相关配置 [qms配置文件设计](https://km.qutoutiao.net/pages/viewpage.action?pageId=182314169)。
	- 平滑重启增加reload_timeout配置控制reload超时时间。
- Fixed
	- 在非qms项目目录执行qms-gen update 也会升级相应依赖。
	
## v1.3.5    2020-04-14
- Fixed
  - 配置多个监听端口+且在容器部署时，实际解析的顺序有可能与预期的不一致。
- 备注
  - ECS多端口部署的没有影响，容器只监听一个端口的服务也没有影响；但如果是监听多个端口的服务+容器部署，请尽快升级！
  
## v1.3.4    2020-04-03
- Changed
  - 支持[一键生成](https://git.qutoutiao.net/gopher/qms/blob/master/docs/development.md#31-%E5%A6%82%E4%BD%95%E5%AE%9E%E7%8E%B0http%E6%9C%8D%E5%8A%A1grpc%E6%9C%8D%E5%8A%A1) http+grpc服务（一次实现，二种协议）
  - IDL方式生成API，支持追加接口

## v1.3.3    2020-03-30
- Fixed
  - 程序的swagger注释没有写tag时，make swagger会panic
  
## v1.3.2    2020-03-27
- Changed
  - 丰富脚手架模板，支持一键生成"完整的"http服务（qms-gen newhttp myapp）
  - 支持IDL的方式，根据proto定义一键生成api [如何使用?](https://git.qutoutiao.net/gopher/qms/blob/master/docs/development.md#34-%E5%A6%82%E4%BD%95%E9%80%9A%E8%BF%87idl%E7%9A%84%E6%96%B9%E5%BC%8F%E6%9D%A5%E7%94%9F%E6%88%90api)
  - 自动集成API文档页面，[如何使用?](https://git.qutoutiao.net/gopher/qms/blob/master/docs/swagger.md), [示例](http://172.25.23.176:12000/apidoc/index.html)
  - 添加统一的qms.Context
  - 支持ctx.RequestID()获取requestid，log输出自动添加request_id，requestid采用traceid
  - 框架内部输出的metrics(QPS/RT)，支持使用方添加自定义分类标签
  - 熔断配置支持配置域名、ipport
- Fixed
  - log输出显示转义字符的问题
  - gorm的debug log开关
- 备注
  - [测试报告](https://km.qutoutiao.net/pages/viewpage.action?spaceKey=INFRA&title=v1.3.2)
  - 通过"qms-gen update"来更新


## v1.3.1
- Changed
  - 支持通过注释生成API文档，并在yapi.qutoutiao.net中统一展示
  - 访问上游的接口，支持指定超时参数(rest.WithTimeout(xx))
  - 访问上游时，支持通过配置ipport列表来自定义路由，在consul/sidecar都异常时，可以作为快速切换的兜底方式
  - access log的写文件方式默认采用异步+循环队列方式（可配置）
  - access log支持显示源IP
  - 发布脚本自动检测consul-agent是否存在
  - 优雅关闭设置10s超时时间
- Fixed
  - 调用上游的重试逻辑没有生效
  - 配置了grpc监听端口，但没有实现grpc服务时，程序启动会panic
  - consul访问不通时，有缓慢goroutine泄漏的问题
- 测试报告
  - [KM地址](https://km.qutoutiao.net/display/INFRA/v1.3.1)

## v1.3.0
- CD发布命令统一为reload，再也不用关心什么时候需要restart发布一次了（之前的版本在监听端口/reload端口/父进程代码变更时，需要restart一次）(注：需要采用qms-gen update来更新本次版本)
- 服务注册时，框架内部把服务名中的"_"转换为"-"（与sidecar保持统一）
- 支持grpc一个监听端口对应多个services
- 新增rest.PostJson/rest/ContextPostJson接口
- 去除框架中自动添加的header头
- 支持CI设置中配置多个cd_app_id配置多个服务的场景，例如:cd_app_id: serverA,serverB,serverC
- 调整部分目录结构，让接口分类更清晰
- 建议：使用方采用qms-gen update的方式更新至v1.3.0

## v1.2.9
- 支持pg演练环境
- 优化平滑重启，先起新子进程，在关闭旧子进程，即使业务程序启动时间长，reload也足够平滑
- reload发布或配置热更新时，如果新版本启动失败，自动回滚到原有版本，并提示本次发版失败，提高容错能力
- 备注：由于本次更新了master/slave中的master部分代码，所以更新后的首次发布需要采用restart方式（sh run.sh restart）

## v1.2.8
- rest调用上游接口，支持添加header参数
- rest调用上游时放开对https的限制
- 修复框架一键更新(qms-gen update)的bug
- 新增配置qms.service.registry.registerDisabled来控制是否只禁用注册（允许发现）
- 整理框架自身的log输出

## v1.2.7
- 增加一键更新框架及生成文件的命令：qms-gen update
- 优化run.sh脚本，提升容错性

## v1.2.6
- 优化trace 使其使用B3头，兼容老服务
- 编译脚本自动收集编译时间、git tag、git commit id，并在程序启动时输出到log
- gin升级到v1.5.0
- 优化metrics中含params的API的path收敛

## v1.2.5
- 支持使用方自定义LoadBalancer
- 优化rest/grpc调用上游的方式，支持options
- 优化部署脚本(run.sh)，不满足平滑重启条件时，reload明确提示失败
- 修复发版或热更新时，metrics数据有概率没有收集的问题
- metrics收集统一收集到thanos数据源

## v1.2.4
- pkg/redis库兼容context参数为nil
- /ping接口同时支持GET&HEAD
- 优化rest/grpc的调用姿势

## v1.2.3
- /ping接口按运维规范返回"OK"
- /metrics自动收集到http://thanos.qtt6.cn/

## v1.2.2
- 修复log输出level配置不生效的问题

## v1.2.1
- 支持输出多个log文件
- 修复redis未配置时panic的问题

## v1.2.0
- 命令行解析兼容subcommand
- 修复配置目录读取依赖于import顺序的问题
- 收敛metrics中的path
- 限制metrics的行数上限
- 统一构建命令为make artifacts，根据是否有vendor目录自动选择模式

## v1.1.20
- 修复由于path过多导致metrics/circuit占用资源过多的问题

## v1.1.17
- 解决path参数过多导致metrics数据过大的问题
- gorm增加初始判断接口，接口调整为返回官方的*gorm.DB
- access.log支持以独立的文件输出

## v1.1.16
- 增加访问mysql的metrics数据
- 支持平滑重启开关的配置项
- 优化run.sh脚本的reload

## v1.1.15
- invoke调用上游支持线程安全
- 访问redis时，没有trace的情况下也正常输出metrics数据

## v1.1.14
- 解决go mod依赖报错的问题
- 支持配置reload的端口

## v1.1.12
- 修复reload时的路径问题

## v1.1.11
- 支持provider端的api级别限流
- 支持通过http方式来触发reload

## v1.1.10
- 规范限流、熔断触发时的状态码

## v1.1.9
- 配置文件中支持全局开关sidecar
- make run增加proxy代理

## v1.1.8
- 支持访问远端的metrics数据
- 支持业务方输出自定义的metrics数据
- 支持输出访问redis的metrics数据
- qms.Init()可以指定conf目录（单元测试可能用到）
- 简化CI构建命令

## v1.1.7
- 支持自动收集metrics数据到promethues，无需额外提工单；
- 限流过滤/ping,/metrics;
- 优化命令行参数；
- 提供是否在容器里的运行时判断；
- 优化ping，log；

## v1.1.6
- 支持部分远端请求走sidecar，部分不走；
- 优化CD运行脚本；

## v1.1.5
- fix脚手架的bug。

## v1.1.4
- 优化脚手架并补充文档。
- 增加获取配置数据的示例。
- 增加健康检测的ping接口。

## v1.1.3
- 增加生成基于qms的脚手架。
- 增加pkg/qenv获取当前环境信息。

## v1.1.2
- 优化redis库的使用方式。

## v1.1.1
- http/grpc皆支持平滑重启。
- 修正默认超时时间。

## v1.1.0
- 明确外部可访问模块与qms内部模块。
- 简化http远程调用方式。

## v1.0.7
- 支持重试条件的自定义配置。
- metrics输出使用方的error-log数目。

## v1.0.6
- 提供errors库，支持error wrap及error输出堆栈信息。
- 提供获取配置文件路径的接口。

## v1.0.5
- 限流实现由阻塞式改为否决式
- 支持k8s部署

## v1.0.4
- fix provider限流时，返回错误码不准确的问题
- 增加grpc的recover处理

## v1.0.3
- 与新CICD结合部署更顺畅
- 支持开启access.log

## v1.0.2
- 解决下载依赖包慢或timeout的问题；
- 支持http服务的平滑重启；

## v1.0.1
- 统一采用-c=/path/to/conf/来指定配置目录；
- 新增用于新CD的make命令，增加run.sh脚本；

## v1.0.0
- 无需额外代码，即可具备全套微服务治理能力；
- 支持快速搭建服务框架；
- 完全支持gin实现http服务；
- 支持grpc服务与调用，方式简洁；
- 开箱即用的log、redis、gorm，且自动实现了trace链；