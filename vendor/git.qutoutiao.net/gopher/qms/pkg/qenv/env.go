package qenv

import (
	"strings"
	"sync"

	"git.qutoutiao.net/gopher/qms/pkg/conf"
)

// ----------------------------------------
//  受支持的运行环境
// ----------------------------------------
const (
	DEV        Env = "dev" // 开发环境
	QA         Env = "qa"  // 测试环境
	MT         Env = "mt"  // 多测环境
	PG         Env = "pg"  // 演练环境
	PRE        Env = "pre" // 预发环境
	PRD        Env = "prd" // 生产环境
	UNKNOWNENV Env = "unknown"

	env_key = "qms.service.env" // 服务当前环境key
)

// ----------------------------------------
//  运行环境类型定义
// ----------------------------------------
type Env string

var (
	env  Env
	once sync.Once
)

// 获取全局运行环境[默认为DEV]
func Get() Env {
	once.Do(func() {
		environment := conf.GetString(env_key, conf.GetString("service.environment", DEV.String())) // 兼容
		env = ToEnv(environment)
	})

	return env
}

// 获取全局运行环境字符串形式
func (e Env) String() string {
	return string(e)
}

// 检查全局运行环境是否受支持并有效
func (e Env) IsValid() bool {
	switch e {
	case DEV, QA, MT, PG, PRE, PRD:
		return true
	default:
		return false
	}
}

func ToEnv(env string) Env {
	env = strings.ToLower(env)
	switch Env(env) {
	case DEV, QA, MT, PG, PRE, PRD:
		return Env(env)
	default:
		return UNKNOWNENV
	}
}

// 检查当前全局运行环境是否是给定的值
func (e Env) Is(env Env) bool {
	return e == env
}

// 检查当前全局运行环境是否是在给定的值范围内
func (e Env) In(envs ...Env) bool {
	for i, j := 0, len(envs); i < j; i++ {
		if e.Is(envs[i]) {
			return true
		}
	}
	return false
}

// 检查当前全局运行环境是否是开发环境
func (e Env) IsDev() bool {
	return e.Is(DEV)
}

// 检查当前全局运行环境是否是测试环境
func (e Env) IsQa() bool {
	return e.Is(QA)
}

// 检查当前全局运行环境是否是多测环境
func (e Env) IsMt() bool {
	return e.Is(MT)
}

// 检查当前全局运行环境是否是演练环境
func (e Env) IsPg() bool {
	return e.Is(PG)
}

// 检查当前全局运行环境是否是预发环境
func (e Env) IsPre() bool {
	return e.Is(PRE)
}

// 检查当前全局运行环境是否是生产环境
func (e Env) IsPrd() bool {
	return e.Is(PRD)
}

// 设置全局运行环境
//func SetEnv(env Env) error {
//	if env.IsValid() {
//		gEnvironment = env
//		return nil
//	} else {
//		return fmt.Errorf("the given environment %s is invalid", env)
//	}
//}

// 设置全局运行环境（字符串自动解析）
//func SetEnvString(env string) error {
//	return SetEnv(Env(strings.ToLower(env)))
//}
