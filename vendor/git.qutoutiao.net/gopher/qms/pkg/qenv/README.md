# pkg/qenv

qms运行环境组件

## import
```bash
import "git.qutoutiao.net/gopher/qms/pkg/qenv"
```

## 示例

```go
// 获取当前运行的环境
env := qenv.Get()

// 判断环境是否有效
if env.IsVaild() {
	
}

// 打印环境
qlog.Info(env)

// 将string转换为环境
qenv.ToEnv("dev")

// 判断属于某个环境
env.IsDev()
env.IsQa()
...

switch env {
    case qenv.QA:
    	xxx
    case qenv.PRD:
    	xxx
}
```