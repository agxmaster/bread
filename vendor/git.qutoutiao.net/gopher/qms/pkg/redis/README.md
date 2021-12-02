# pkg/redis

## 功能简介
- 开箱即用地拿到go-redis的Client，自动从配置文件解析配置初始化go-redis，使用方不需要再次读取配置，传递参数;
- 自动集成了trace;
- 自动把访问redis的统计数据输出到metrics;

## 使用示例:
- 参考代码
```go
import(
    qmsredis "git.qutoutiao.net/gopher/qms/pkg/redis"
    "github.com/go-redis/redis/v7"
)

//...
//程序初始化阶段，判断redis配置及网络连接的正确性.
if err := qmsredis.CheckValid(); err!=nil {
    panic(err)
}

//...

redisName := "instance1" //配置中的redis名称
ctx := xxx           //上下文传递的context.Context
redisCli := qmsredis.ClientWithTrace(ctx, redisName)         //【方式一】取到go-redis的Client对象
//redisCli := qmsredis.Client(redisName).WithContext(ctx)    //【方式二】取到go-redis的Client对象
//...                //开始使用go-redis
```

- 参考配置(app.yaml)
```yaml
redis:
  instance1:  #redis实例名称，作为key用于表示某个redis实例，名字自定义取, 例如: local, master, slave, feed_master, ...
    addr: 127.0.0.1:6379  #[MUST]redis地址
    password: ""          #redis密码
    database: 0           #redis db index, {default: 0}
    dial_timeout: 5000    #连接超时时间，单位: millisecond, {default: 5000}
    read_timeout: 1000    #读超时时间，单位: millisecond, {default: 1000}
    write_timeout: 1000   #写超时时间，单位: millisecond, {default: 1000}
    max_retries: 0        #最大重试次数, {default: 0}
    pool_size: 0          #最大连接数大小, {default: runtime.NumCPU*10}
    min_idle_conns: 0     #一直保持的空闲连接数(无论是否有请求),一般为0即可 {default: 0}
```
