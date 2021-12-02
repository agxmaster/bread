# Redis

[![pipeline status](https://git.qutoutiao.net/golib/redis/badges/master/pipeline.svg)](https://git.qutoutiao.net/golib/redis/commits/master)
[![coverage report](https://git.qutoutiao.net/golib/redis/badges/master/coverage.svg)](https://git.qutoutiao.net/golib/redis/commits/master)


基于 `github.com/go-redis/redis/v8` 封装 *Redis* 客户端，主要提供以下功能：

- 单例模式，支持多实例管理功能；

- 标准化 Golang API 模式；

- 支持 Trace 功能；

- 支持 Metrics 功能；


## 安装

```bash
$ go get -v git.qutoutiao.net/golib/redis
```

## 快速开始
1. 配置说明
    ```golang
    type Config struct {
        Network              string `yaml:"network"`                 //网络类型，支持：tcp，unix；默认tcp
        Addr                 string `yaml:"addr"`                    //网络地址，ip:port，如：172.0.0.1:6379
        Passwd               string `yaml:"password"`                //密码
        DB                   int    `yaml:"database"`                //redis database，默认0
        DialTimeout          int    `yaml:"dial_timeout"`            //连接超时时间，默认1000ms
        ReadTimeout          int    `yaml:"read_timeout"`            //socket 读超时时间，默认100ms
        WriteTimeout         int    `yaml:"write_timeout"`           //socket 写超时时间，默认100ms
        PoolSize             int    `yaml:"pool_size"`               //连接池最大数量，默认200
        PoolTimeout          int    `yaml:"pool_timeout"`            //从连接池获取连接超时时间，默认ReadTimeout + 1000ms
        MinIdleConns         int    `yaml:"min_idle_conns"`          //连接池最小空闲连接数，默认30
        MaxRetries           int    `yaml:"max_retries"`             //重试次数，默认0
        TraceIncludeNotFound bool   `yaml:"trace_include_not_found"` //是否将key NotFound 作为错误记录在trace中，默认为否
    }
    ```

2. 初始化
    ```golang
    redis.Register("my-redis-demo", &redis.Config{
        Addr:                 "127.0.0.1:6379",
    })
    ```

3. 使用
    ```golang
    func main() {
    	logger := qulibs.NewLogger(qulibs.LogDebug)
    
    	redisGroupName := "demo-redis"
    
    	// load redis client by group name
    	client, err := redis.GetClient(redisGroupName)
    	if err != nil {
    		logger.Errorf("redis.GetClient(%s): %v", redisGroupName, err)
    		return
    	}
    
    	// set a new key-value pair
    	var (
    		ctx   = context.Background()
    		key   = "redis-key"
    		value = "redis-value"
    	)
    
    	// create an new span for current operation
    	span := opentracing.GlobalTracer().StartSpan("demo-redis")
    	defer span.Finish()
    
    	// inject context with the span
    	ctx = opentracing.ContextWithSpan(ctx, span)
    
    	err = client.Set(ctx, key, value, time.Minute).Err()
    	if err != nil {
    		logger.Errorf("client.Set(%v, %v, time.Minute): %v", key, value, err)
    		return
    	}
    
    	// read a value from redis with the key
    	nval, nerr := client.Get(ctx, key).Result()
    	if nerr != nil {
    		logger.Errorf("client.Get(%v): %v", key, nerr)
    		return
    	}
    
    	logger.Infof("Get value of %v: %v", key, nval)
    }
    ```
    
## 基于配置文件直接初始化
本库支持直接从配置文件初始化，默认支持 2 种配置格式
1. 配置文件说明
    ```yaml
    # 独立配置格式
    # /path/to/redis.yml
    demo-redis:
      network: tcp
      addr: 127.0.0.1:6379

    local-redis:
      network: tcp
      addr: 127.0.0.1:6480
      ...
    ```
    
    ```yaml
    # 组件配置格式
    # /path/to/components.yml
    redis:
      demo-redis:
        network: tcp
        addr: 127.0.0.1:6379

      local-redis:
        network: tcp
        addr: 127.0.0.1:6480
        ... 
        ```
2. 初始化
    ```golang
    func init() {
    	// init redis from config file
    	root, err := os.Getwd()
    	if err != nil {
    		panic(err)
    	}
    
    	filename = filepath.Join(root, "examples", "/path/to/redis.yml")
    
    	err = redis.Init(filename)
    	if err != nil {
    		panic(err)
    	}
    }
    ```
3. 使用
    ```golang
    func main() {
    	logger := qulibs.NewLogger(qulibs.LogDebug)
    
    	redisGroupName := "demo-redis"
    
    	// load redis client by group name
    	client, err := redis.GetClient(redisGroupName)
    	if err != nil {
    		logger.Errorf("redis.GetClient(%s): %v", redisGroupName, err)
    		return
    	}
    
        //使用client...
    }
    ```

## API 中使用 Trace 功能

基于本库提供的 Trace 功能，只需在服务 API 中将 `http.Request.Context()` 值作为参数传入即可实现 Trace 功能。

```golang
func (user *UserService) Login(request *http.Request) {
    redisClient, redisErr := redis.GetClient("user-redis")
    if redisErr != nil {
        // handle error
        return
    }

    result, err := redisClient.Get(request.Context(), ...).Result()
    if err != nil {
        // handle error
        return
    }

    // handle login
```
