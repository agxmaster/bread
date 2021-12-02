# pkg/qlog

## 功能简介
- 开箱即用的log输出
- 支持结构化字段
- 支持json
- 支持输出多个文件

## 示例
```go
import "git.qutoutiao.net/gopher/qms/pkg/qlog"
func main() {
    qlog.Info("xxx")
    qlog.Infof("a=%s,b=%d", "xxx", 123)
    qlog.Error(err)
    qlog.Error(err)
    qlog.WithField("fk", 123).Info("xxx")
}
```

