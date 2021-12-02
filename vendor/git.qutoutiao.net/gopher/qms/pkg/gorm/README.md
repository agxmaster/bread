# pkg/gorm

## 功能简介
- 开箱即用地拿到*gorm.DB，自动从配置文件解析配置、初始化gorm，使用方不需要再次读取配置，传递参数;
- 自动集成了trace;
- 自动集成了metrics;

## 使用示例:
- 参考代码
```go
import(
    qmsgorm "git.qutoutiao.net/gopher/qms/pkg/gorm"
    "github.com/jinzhu/gorm"
)

//...
//程序初始化阶段，判断redis配置及网络连接的正确性.
if err := qmsgorm.CheckValid(); err!=nil {
    panic(err)
}

//...

mysqlName := "local" //配置中的mysql名称
ctx := xxx           //上下文传递的context.Context
gormDB := qmsgorm.ClientWithTrace(ctx, mysqlName)          //取到*gorm.DB
//...            //开始使用gorm
```

- 参考配置（app.yaml）
```yaml
mysql:
  local:  #mysql实例名称，作为key用于表示某个mysql实例，自定义命名, 例如: local, master, slave, feed_master, ...
    dsn: "root:@tcp(127.0.0.1:3306)/information_schema?charset=utf8mb4&parseTime=True&loc=Local"  #[MUST]
    dial_timeout: 5000    #连接超时时间, 单位: millisecond, {default: 5000}
    read_timeout: 5000    #读超时时间, 单位：millisecond, {default: 5000}
    write_timeout: 3000   #写超时时间, 单位：millisecond, {default: 3000}
    max_open_conns: 256   #最大连接数大小, {default: 256}
    max_idle_conns: 10    #最大空闲的连接的个数, {default: 10}
    max_life_conns: 0     #连接的生命时间,超过此时间，连接将关闭后重新建立新的，0代表忽略相关判断,单位:second, {default: 0}
    debug_sql: false      #是否开启debug, {default: false}
```
