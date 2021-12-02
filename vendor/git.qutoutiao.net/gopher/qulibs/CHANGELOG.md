### v1.23.0

- upgrade `discovery` to v1.15.5


### v1.19.9 发布

- 1, qulibs
    - `gorm` 新增 `debug_sql` 配置，当设为 `true` 时将打印所有 SQL 语句；
    - `logger` 优化性能，避免在不满足 Level 的调用中产生 allocs；

### v1.19.8 发布

- 1, qulibs
    - 新增 `trace.Config.EnableSpanPool` 配置属性；
    - 移除 `trace.Config.DisableSpanPool` 配置属性；
    
### v1.19.7 发布

- 1, qulibs
    - 新增 `trace.Config.DisableSpanPool` 配置属性；

### v1.19.2 发布

- 1, qulibs
    - 更新 `trace` 库，支持 1 ~ 10000 的 sample rate；

### v1.19.1 发布

- 1, qulibs
    - 更新依赖包，包括 qudiscovery, grpc, gorm 等；
    - 更新 `gorm` trace 支持，默认以当前节点 IP 为 peer hostname；
    - 修正 `gorm` 默认超时配置问题；

### v1.19.0 发布

- 1，qulibs
    - 重构 `qugrpc` 包为 `grpc` 包，修正一系列问题；
    
### v1.18.0 发布

- 1，qulibs
    - 修正 `logger` 包无法正确应用 *Level* 问题；

### v1.17.0 发布

- 1，qulibs
    - 新增 logger pkg，提供 zerolog 统一封装；
    - 新增 `Logger` 接口定义和 `DummyLogger` 实现
    - 新增 `qulibs/gorm` 封装

### v1.1.1 发布

- 1，Redis
    - 新增 `client.Select(db)` 接口
    - 重构 `Manager.NewClient()` 方法，并添加测试用例
