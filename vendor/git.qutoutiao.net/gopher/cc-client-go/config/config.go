package config

import (
	"io"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"

	"git.qutoutiao.net/gopher/cc-client-go"
)

type configOnceKey struct {
	projectName string
	env         cc.Env
}

type configInitCache struct {
	*cc.ConfigCenter
	err error
}

var (
	configOnce = sync.Map{}
	configInit = sync.Map{}
)

// Configuration 初始化配置中心sdk的配置
type Configuration struct {
	// ProjectName 项目名称，值为配置中心中的项目名称
	ProjectName string `toml:"project_name"`
	// ProjectToken 项目的token，在配置中心http:/cc.qutoutiao.net的项目主页查看
	ProjectToken string `toml:"project_token"`
	// Env 机器所属的环境
	Env cc.Env `toml:"env"`
}

func validEnv(env cc.Env) bool {
	switch env {
	case cc.DEV, cc.QA, cc.PG, cc.PRE, cc.PRD:
		return true
	default:
		return false
	}
}

func NewConfiguration(projectName string, projectToken string, env cc.Env) *Configuration {
	return &Configuration{
		ProjectName:  projectName,
		ProjectToken: projectToken,
		Env:          env,
	}
}

func (c Configuration) NewCC(options ...Option) (*cc.ConfigCenter, io.Closer, error) {
	if !validEnv(c.Env) {
		nullCC := cc.NewNullConfigCenter(c.ProjectName, c.ProjectToken, c.Env)
		return nullCC, nullCC, cc.ErrEnvEnumValue
	}
	val, _ := configOnce.LoadOrStore(configOnceKey{
		projectName: c.ProjectName,
		env:         c.Env,
	}, new(sync.Once))
	var configCenter *cc.ConfigCenter
	var err error
	val.(*sync.Once).Do(func() {
		configCenter, _, err = c.newCC(options...)
		configInit.Store(configOnceKey{
			projectName: c.ProjectName,
			env:         c.Env,
		}, configInitCache{
			ConfigCenter: configCenter,
			err:          err,
		})
	})
	val, ok := configInit.Load(configOnceKey{
		projectName: c.ProjectName,
		env:         c.Env,
	})
	if !ok {
		return nil, nil, errors.New("配置中心初始化失败")
	}
	ccCache := val.(configInitCache)
	return ccCache.ConfigCenter, ccCache.ConfigCenter, ccCache.err
}

func (c Configuration) newCC(options ...Option) (*cc.ConfigCenter, io.Closer, error) {
	ip, err := cc.HostIP()
	if err != nil {
		nullCC := cc.NewNullConfigCenter(c.ProjectName, c.ProjectToken, c.Env)
		return nullCC, nullCC, errors.WithStack(err)
	}
	opts := applyOptions(options...)
	// opts.appLoggers[1] = cc.NewAppLogger(client, &sdkInfo)
	// opts.diagnosticLogger[1] = remote.NewLogger(client, &sdkInfo)
	ccOptions := []cc.Option{
		cc.DevConfigFilePath(opts.DevConfigFilePath),
		cc.DiagnosticLogger(opts.diagnosticLogger),
		cc.QAServerURL(opts.QAServerURL),
		//cc.AppLoggerOpt(opts.appLoggers),
		cc.ClientIP(ip.String()),
		cc.RestoreWhenInitFail(opts.restoreWhenInitFail),
		cc.RetryableCodes([]codes.Code{codes.ResourceExhausted, codes.Unavailable, codes.Canceled}),
		cc.OnChange(opts.onChange),
		cc.BackupDir(opts.backupDir),
		cc.MaxInterval(opts.maxInterval),
		cc.InitialInterval(opts.initialInterval),
		cc.JitterFraction(opts.jitterFraction),
		cc.DebugOpt(opts.Debug),
	}
	configCenter, err := cc.NewConfigCenter(c.ProjectName, c.ProjectToken, c.Env, ccOptions...)
	return configCenter, configCenter, err
}
