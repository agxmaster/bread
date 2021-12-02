# Gorm 

[![pipeline status](https://git.qutoutiao.net/golib/gorm/badges/master/pipeline.svg)](https://git.qutoutiao.net/golib/gorm/commits/master)
[![coverage report](https://git.qutoutiao.net/golib/gorm/badges/master/coverage.svg)](https://git.qutoutiao.net/golib/gorm/commits/master)


基于 `gorm.io/gorm` (原 `github.com/jinzhu/gorm`) 封装 *MySQL* 客户端，主要提供以下功能：

- 单例模式，支持多实例管理功能；

- 支持 Trace 功能；

- 支持 Metrics 功能；

- 支持慢日志功能，默认 > 100ms；

## 安装

```bash
$ go get -v git.qutoutiao.net/golib/gorm
```


## 快速开始

1. 配置说明 

    ```golang
    type Config struct {
        Driver                 string        `yaml:"driver"`                    // 数据库类型，当前支持：mysql，postgres，sqlite，默认 mysql
        DSN                    string        `yaml:"dsn"`                       // 数据库 DSN
        DialTimeout            time.Duration `yaml:"dial_timeout"`              // 连接超时时间，默认 1000ms
        ReadTimeout            time.Duration `yaml:"read_timeout"`              // socket 读超时时间，默认 3000ms
        WriteTimeout           time.Duration `yaml:"write_timeout"`             // socket 写超时时间，默认 3000ms
        MaxOpenConns           int           `yaml:"max_open_conns"`            // 最大连接数，默认 200
        MaxIdleConns           int           `yaml:"max_idle_conns"`            // 最大空闲连接数，默认 80
        MaxLifetime            int           `yaml:"max_life_time"`             // 空闲连接最大存活时间，默认 600s
        TraceIncludeNotFound   bool          `yaml:"trace_include_not_found"`   // 是否将NotFound error作为错误记录在trace中，默认为否
        MetricsIncludeNotFound bool          `yaml:"metrics_include_not_found"` // 是否将NotFound error作为错误记录在metrics中，默认为否
        DebugSQL               bool          `yaml:"debug_sql"`                 // 开启 SQL 调试模式，即输出所有 SQL 语句
    }
    ```
    
2. 初始化

    ```golang
    func init() {
        gorm.Register("my-demo-mysql", &gorm.Config{
            Driver: "mysql",
            DSN:    "root:@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
        })
    }
    ```
    
3. 使用

    ```golang
    type DemoModel struct {
        *gormio.Model

        Subject string
    }

    func main() {
        groupName := "my-demo-mysql"
        logger := qulibs.NewLogger(qulibs.LogDebug)

        client, err := gorm.GetClient(groupName)
        if err != nil {
            logger.Errorf("gorm.GetClient(%s): %v", groupName, err)
            return
        }

        ctx := context.Background()

        // create an new record
        tm := &DemoModel{
            Subject: "Demo Subject",
        }

        result := client.Save(ctx, tm)
        if result.Error != nil {
            logger.Errorf("client.Save(%+v): %v", tm, result.Error)
            return
        }

        // query record by id
        var tmpm DemoModel

        result = client.Where("id", tm.ID).First(ctx, &tmpm)
        if result.Error != nil {
            logger.Errorf("client.First(id=%v): %v", tm.ID, result.Error)
            return
        }
        logger.Infof("Retrieved record: %+v", tmpm)

        // update record by subject
        result = client.Model(new(DemoModel)).Where("subject", tm.Subject).Update(ctx, "subject", "New Demo Subject")
        if result.Error != nil {
            logger.Errorf("client.Update(subject=%s): %v", tm.Subject, result.Error)
            return
        }
    }
    ```

## 基于配置文件直接初始化

`gorm` 封装提供从配置文件初始化的简易操作，默认支持 2 种配置格式。

1. 配置文件示例

    ```yaml
    # 独立配置格式
    # filename = /path/to/gorm.yml
    my-demo-mysql:
      driver: mysql
      dsn: root:@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local
    my-second-mysql:
      driver: mysql
      dsn: root:@tcp(localhost:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
    ```
   
    ```yaml
    # 组件配置格式
    # filename = /path/to/components.yml
    mysql:
      my-demo-mysql:
        driver: mysql
        dsn: root:@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local
      my-second-mysql:
        driver: mysql
        dsn: root:@tcp(localhost:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
    redis:
      ....
    ```
  
2. 初始化

    ```golang
    func init() {
        // 加载配置文件
        filename := "/path/to/components.yml" 
 
        err := gorm.Init(filename)
        if err != nil { 
            panic(err.Error())
        }
    }
    ```
    
3. 使用

    ```golang
    type DemoModel struct {
        *gormio.Model

        Subject string
    }

    func main() {
        groupName1 := "my-demo-mysql"
        groupName2 := "my-second-mysql"
        logger := qulibs.NewLogger(qulibs.LogDebug)

        client1, err := gorm.GetClient(groupName1)
        if err != nil {
            logger.Errorf("gorm.GetClient(%s): %v", groupName1, err)
            return
        }
        
        client2, err := gorm.GetClient(groupName2)
        if err != nil {
            logger.Errorf("gorm.GetClient(%s): %v", groupName2, err)
            return
        }
        
        // 使用 client...
    }
    ```

## Metrics 说明

`gorm` 封装提供了以下核心 metrics 监控。

- `gorm_op_totals`：操作 QPS

- `gorm_op_failures`：操作失败 QPS

- `gorm_op_latency_sum`: 操作延时
