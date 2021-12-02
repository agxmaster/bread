# 简介
qms(Qtt MicroService)是由基础架构部开发的一款Golang微服务开发框架，目标是让业务方在打造微服务架构时，只需关心业务逻辑开发，自动具备服务治理能力，以提升业务开发效率。
![qms-design](docs/images/design.png)
- 名词解释
    - qms框架：qms-gen + qms库 + protoc-gen-qms插件
    - qms库: 封装了程序初始化(qms.Init)、运行(qms.Run)等方法，库内部实现了丰富的服务治理功能，qms/pkg下集成了许多有助于提升开发效率的子库。
    - qms-gen: 脚手架生成器，可以生成基于qms的业务服务。也可以用于更新qms。
    - protoc-gen-qms: proto生成*.pb.go的自定义插件。
    - Makefile: 编译构建脚本（用于CI）
    - run.sh: 运行脚本（用于CD）
    - 上游: 上游服务，指被依赖的服务，例如:A->B，B为上游服务
    
# 使用qms的收益
- **开发环节**    
  - 支持一键生成http+grpc服务，支持多协议、多端口；
  - 支持IDL方式一键生成API；
  - 集成公司consul，自动具备服务注册、服务发现能力；
  - 集成公司配置中心，支持配置热更新；
  - 集成公司trace，自动具备链路追踪能力；
  - 自动生成服务指标的metrics数据，包括服务入口、访问上游服务/redis/mysql/nsq、error-log数目、goroutines等。也支持使用方自定义metrics数据；
  - 自动具备限流能力，支持服务入口限流、也支持访问上游服务的限流，支持服务级别、也支持API级别；
  - 自动具备多种隔离容错手段，包括：超时、熔断、快速开关等；
  - 自动具备优雅重启、优雅关闭能力；
  - 自动兼容sidecar代理，也支持在sidecar异常时一键切换备用方案；
  - 自动兼容容器部署；
  - qms/pkg下封装了多种基础库，有助于提升使用方的开发效率；
- **部署环节**    
  - 简化CI配置，只需填写paas项目id;
  - 简化CD配置，参数自动填充；
- **监控告警**    
  - metrics数据自动收集；
  - 监控页面自动创建，服务指标、系统指标、容器指标的监控曲线自动就绪；
  - 告警规则自动创建；
- **公司规范**    
  - 满足[公司运维规范](https://km.qutoutiao.net/pages/viewpage.action?pageId=96768319)
  - 满足[公司高可用规范](http://km.qutoutiao.net/pages/viewpage.action?pageId=150739202)

# Get Started
- 推荐Go版本: >= 1.13

## 快速搭建应用
1.设置代理    
```bash
#Go version >= 1.13
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE="*.qutoutiao.net"
```
备注: 代理是为了解决墙的问题，上述代理只在module-ware模式生效，所以非go.mod目录下执行go get时，前面需要指定GO111MODULE=on

2.准备条件
- 安装qms-gen（生成器）    
```bash
GO111MODULE=on go get git.qutoutiao.net/gopher/qms/cmd/qms-gen
```
- 安装protoc-gen-qms (protoc插件)    
```bash
GO111MODULE=on go get git.qutoutiao.net/gopher/qms/cmd/protoc-gen-qms
```
- 安装protoc (protobuf解析器)
  - Mac
  ```bash
  ruby -e "$(curl -fsSL  https://raw.githubusercontent.com/Homebrew/install/master/install)" #如果已安装brew则跳过该步骤
  brew install protobuf
  ```
  - Windows    
  下载[Protobuf安装程序](https://github.com/protocolbuffers/protobuf/releases)，将解压后的 bin 目录添加至环境变量「path」中, 如：D:\protobuf\protoc-3.11.4-win64\bin

3.生成应用
- 生成最简的HTTP应用
```bash
qms-gen new myapp      //生成最简的http服务
```
- 生成"完整的"应用（涵盖目录组织结构、log、api、api文档、redis、mysql等）
```bash
qms-gen newhttp myapp  //生成http服务
qms-gen newgrpc myapp  //生成grpc服务
qms-gen newboth myapp  //生成http+grpc服务（一次实现，两种协议）
```

## 参考文档: 
  - [qms框架开发文档](https://git.qutoutiao.net/gopher/qms/blob/master/docs/development.md)
  - [qms框架功能清单](https://km.qutoutiao.net/pages/viewpage.action?pageId=182308324)
  - [历史版本的问题与解决](https://km.qutoutiao.net/pages/viewpage.action?pageId=150747135)
  - [qms框架的metrics规范](https://km.qutoutiao.net/pages/viewpage.action?pageId=173535986)
  
# FAQ
  - [FAQ](https://git.qutoutiao.net/gopher/qms/blob/master/docs/FAQ.md)
  - 任何问题与建议，请联系王冀航，杨毅鹏



