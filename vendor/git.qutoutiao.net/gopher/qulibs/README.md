# qulibs

[![pipeline status](https://git.qutoutiao.net/gopher/qulibs/badges/master/pipeline.svg)](https://git.qutoutiao.net/gopher/qulibs/commits/master) [![coverage report](https://git.qutoutiao.net/gopher/qulibs/badges/master/coverage.svg)](https://git.qutoutiao.net/gopher/qulibs/commits/master)

趣星人推荐 Golang 开发基础库封装。

## 注意！！！

qulibs 下所有以 `quXXX` 命名的包均为旧版本实现，请选择使用其它对应的包！！！

## 开发规范

- 仅提供实例对象管理或连接池管理功能！如果第三方库已包含连接池功能，则禁止再次封装新的实现，除非你确定第三方库实现有问题；

- 不提供业务相关的特性增强功能！如数据库分库分表，数据库主从读写分离等；

- 推荐实现 `Logger` 接口注入逻辑，默认使用 `qulibs.DummyLogger` 实现；

- 推荐包含 `Trace` 服务支持（具体接口规范待确定）；

- 推荐包含 `Metrics` 服务支持（Prometheus，具体接口规范待确定）；

### 新增基础库

- fork 或创建特性开发分支

- 实现基础库封装，封装代码结构约定如下

    ```bash
    redis/
    ├── README.md        # 必选，实现说明和使用说明；
    ├── client.go        # 必选，单实例模式封装实现，建议使用组合方式扩展功能；
    ├── client_test.go   # 必选，单实例模式扩展功能测试用例；
                         #      建议测试覆盖 empty input/invalid empty/valid input 场景，并包含并发场景；
    ├── config.go        # 必须，封装实现配置定义，建议每个封装实现定义扩展配置；
                         #      如果第三方库包含配置对象，则建议使用组合方式扩展定义；
    ├── config_test.go   # 可选，扩展配置测试用例；
    ├── errors.go        # 必选，基础库返回的通用错误定义；
    ├── examples         # 可选，基础库示例代码目录；
    │   └── main.go
    ├── init.go          # 可选，如果基础库包含全局对象，则必须定义在此文件中的 func init() {} 初始化函数中；
    ├── manager.go       # 可选，多实例模式封装实现，必须基于 client.go 单示例模式封装对象；
    └── manager_test.go  # 可选，多实例模式扩展功能测试用例；
    ```

- 更新 `vendor` （注意：vendor 为 git submodule，repo：git@git.qutoutiao.net:gopher/qulibs-golang.git）

- 更新 CHANGELOG.md

- 提交 PR

- 创建新 tag

### 更新基础库

- fork 或创建开发分支

- 添加或更新 API，并确保 API 向前兼容性

- 更新 `vendor` （注意：vendor 为 git submodule，repo：git@git.qutoutiao.net:gopher/qulibs-golang.git）

- 更新 CHANGELOG.md

- 提交 PR

- 创建新 tag
