package runtime

import (
	"io/ioutil"
	"os"
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

//Status
const (
	StatusRunning = "UP"
	StatusDown    = "DOWN"
)

//编译时的基础信息
var (
	BuildTime   string //当前版本的编译时间
	GitCommitID string //当前版本的git commit id
	GitTag      string //当前版本的git tag
)

//HostName is the host name of service host
var HostName string

//ServiceID is the service id in registry service
var ServiceID string

//ServiceName represent self name
var ServiceName string

//Environment is usually represent as development, testing, production and  acceptance
var Environment string

//Schemas save schema file names(schema IDs)
var Schemas []string

//App is app info
var App string

//Version is version info
var Version string

//MD is service metadata
var MD map[string]string

//InstanceMD is instance metadata
var InstanceMD map[string]string

//InstanceID is the instance id in registry service
var InstanceID string

//InstanceStatus is the current status of instance
var InstanceStatus string

// InsideDocker is the flag if current process running in docker
var InsideDocker bool

type Service struct {
	Service     string
	Environment string
	Version     string
}

//Init initialize runtime info
func Init(service *Service) (err error) {
	ServiceName = service.Service // 暂时用ServiceName 兼容
	Version = service.Version
	Environment = service.Environment
	App = common.DefaultApp
	if HostName == "" {
		HostName, err = os.Hostname()
		if err != nil {
			qlog.Error("Get hostname failed:" + err.Error())
			return
		}
	}

	qlog.Infof("service.name is %s", ServiceName)
	qlog.Infof("service.environment is %s", Environment)
	qlog.Infof("service.version is %s", Version)
	qlog.Infof("Hostname is %s", HostName)
	qlog.Infof("runtime.BuildTime=%s", BuildTime)
	qlog.Infof("runtime.GitTag=%s", GitTag)
	qlog.Infof("runtime.GitCommitID=%s", GitCommitID)

	inDocker := hasCGroupDocker()
	if inDocker {
		qlog.Info("running in DOCKER")
	}
	InsideDocker = inDocker
	return nil
}

func hasCGroupDocker() bool {
	bytes, err := ioutil.ReadFile("/proc/self/cgroup")
	if err != nil {
		return false
	}
	cGroupContent := string(bytes)
	return strings.Contains(cGroupContent, "docker")
}
