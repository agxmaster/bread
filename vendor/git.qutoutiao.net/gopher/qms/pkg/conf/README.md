# pkg/conf

## 功能简介
- qms.Init()后，conf/目录下的配置文件（app.yaml/advanced.yaml/log.yaml）数据自动读取到了内存中；
- 使用方自己的添加的配置，需要放置在app.yaml文件中；
- 使用方可以调用conf.Unmarshal()，将所有配置数据一次性读取到自定义的数据结构中；
- 也可以获取单个配置字段的数据，例如: conf.Get(), conf.GetString(),...;
- 也可以判断某个配置字段是否存在，例如: conf.Exist();

## 如何使用
### 读取多个配置字段
- 配置
```yaml
#app.yaml

modelA:
  property1: value1
  property2: value2
modelBs:
  key1:
    property1: value1
    property2: value2
  key2:
    property1: value1
    property2: value2
```
- 代码
```go
import(
	"git.qutoutiao.net/gopher/qms/pkg/conf"
)

type Config Struct {
	A StructA `yaml:"modelA"`
	Bs map[string]StructB `yaml:"modelBs"`
}
// 获取配置
var cfg Config
if err := conf.Unmarshal(&cfg); err != nil {
    qlog.Error(err)
}
```

### 读取单个配置字段
- 配置
```yaml
#app.yaml

modelA:
  key1:
    property1: value1
```
- 代码
```go
import(
	"git.qutoutiao.net/gopher/qms/pkg/conf"
)

// 判断是否存在
exist := conf.Exist("modelA.key1.property1")

// 读取配置
strValue := conf.GetString("modelA.key1.property1", "defaultValue")

```

## 参考示例
- [examples/config](https://git.qutoutiao.net/gopher/qms/tree/master/examples/config)