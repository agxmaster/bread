###### 配置中心golang SDK

## 安装
```
go get -u  git.qutoutiao.net/gopher/cc-client-go
```
## 快速上手
参考example中的例子[example/main.go](./example/main.go)
## 单例形式进行初始化
```
import (
	"git.qutoutiao.net/gopher/cc-client-go"
	"git.qutoutiao.net/gopher/cc-client-go/config"
)
projectName, projectToken := "xx", "xx"
cfg := config.NewConfiguration(projectName, projectToken, cc.QA)
c, _, err := cfg.NewCC
// 将配置中心对象设置为全局对象
cc.SetGlobalCC(c)
// 获取key, 支持返回值的类型为str, int, float, bool
strVal := cc.GetString("key", "")
intVal := cc.GetInt("intKey", 0)
floatVal := cc.GetFloat("floatKey", 0.0)
boolVal := cc.GetBool("boolKey", false)

```
## 初始化选项
配置中心初始化选项包括备份文件路径、回调函数、是否开启debug模式、开发环境读取的本地路径配置文件

### 开启debug
```go
package demo

import (
	"git.qutoutiao.net/gopher/cc-client-go"
	"git.qutoutiao.net/gopher/cc-client-go/config"
)

projectName, projectToken := "xx", "xx"
cfg := config.NewConfiguration(projectName, projectToken, cc.QA)
c, closer, err := cfg.NewCC(config.Debug(true))

```
### 自定义备份文件路径
sdk启动时会从本地的备份文件还原，并将备份文件的版本号请求配置中心服务，如果已经是最新的版本号，配置中心服务则不返回内容，否则返回最新发布的版本号，sdk加载到内存中后会备份到本地，同时用户在配置中心平台手动发布，sdk接收到配置，加载到内存中后也会备份到本地。
```go
c, closer, err := cfg.NewCC(config.backupDir("/data/etc/cc/sdk/"))
```
### 回调函数
当配置更新后，会执行用户配置的回调函数，函数的签名为
```go
func (*config.ConfigCenter) error
```
配置回调函数
```go
c, closer, err := cfg.NewCC(config.OnChange(func(c *config.ConfigCenter) error {
    fmt.Printf("on change, new key values: %v\n", c.GetAll())
    return nil
}))
```
### 开发环境使用sdk
目前配置中心支持的环境为qa, pg, pre, prd，对于dev环境，为封装统一的接口，可以通过本地文件作为数据源，文件的格式为Java的properties格式。键值对可以通过配置中心的某个环境拷贝过来然后修改。初始化的时候指定读取的路径
```
projectName, projectToken := "xx", "xx"
cfg := config.NewConfiguration(projectName, projectToken, cc.DEV)
c, closer, err := cfg.NewCC(config.Debug(true),config.DevConfigFilePath("/tmp/cc/sdk/xx.properities"))
```

## 关闭配置中心
构造配置中心对象的时候会返回io.Closer实例，推荐在main函数执行完毕之前调用Close方法来关闭与配置中心的连接以及刷日志。
```
c, closer, err := cfg.NewCC()
defer closer.Close()
```
## 文件降级
如果配置中心服务不可用，任可以通过修改本地的文件来修改配置，目录为数据备份的目录，默认为: /data/etc/cc/sdk/，数据文件名称为：项目名称_环境.json
1. 修改数据文件，config_variable_key为配置的key，config_variable_value为配置的value
2. 删除checksum文件，sdk会跳过数据的checksum，使得修改生效，checksum的文件名称为：项目名称_环境.checksum