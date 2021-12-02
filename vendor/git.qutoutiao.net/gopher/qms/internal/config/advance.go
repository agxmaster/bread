package config

import (
	"fmt"
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/fileutil"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
)

type Registry struct {
	Disabled bool
	Address  string
	Pilot    string
	CacheDir string
}

func (r *Registry) init(qconf *qconf.Qconf) {
	// TODO: 如果需要服务发现和注册 则必须验证address
	r.Disabled = qconf.GetBool("qms.registry.disabled", qconf.GetBool("qms.service.registry.disabled", qconf.GetBool("qms.service.registry.registrator.disabled")))
	r.Address = qconf.GetString("qms.registry.address", qconf.GetString("qms.service.registry.address", getDefaultStringByEnv("registry.address")))
	r.Pilot = qconf.GetString("qms.registry.pilot", qconf.GetString("qms.service.registry.pilot", getDefaultStringByEnv("registry.pilot")))
	r.CacheDir = qconf.GetString("qms.registry.cache_dir", qconf.GetString("qms.service.registry.cacheDir", defaultRegistryCacheDir))
}

type Metrics struct {
	Enabled          bool
	Path             string
	RuntimeDisabled  bool
	CircuitDisabled  bool
	ErrorLogDisabled bool
	RedisDisabled    bool
	GormDisabled     bool
	AutoMetrics      AutoMetrics // TODO: 和毅鹏讨论下是否需要或只保留enabled
}

func (m *Metrics) init() {
	m.Enabled = qconf.GetBool("qms.metrics.enabled")
	m.Path = qconf.GetString("qms.metrics.path", qconf.GetString("qms.metrics.apiPath", constutil.DefaultMetricPath))
	m.RuntimeDisabled = qconf.GetBool("qms.metrics.runtime_disabled", qconf.GetBool("qms.metrics.disableGoRuntimeMetrics"))
	m.CircuitDisabled = qconf.GetBool("qms.metrics.circuit_disabled", qconf.GetBool("qms.metrics.disableCircuitMetrics"))
	m.ErrorLogDisabled = qconf.GetBool("qms.metrics.error_log_disabled", qconf.GetBool("qms.metrics.disableErrorLogMetrics"))
	m.RedisDisabled = qconf.GetBool("qms.metrics.redis_disabled", qconf.GetBool("qms.metrics.disableRedisMetrics"))
	m.GormDisabled = qconf.GetBool("qms.metrics.gorm_disabled", qconf.GetBool("qms.metrics.disableGormMetrics"))
	m.AutoMetrics.init(qconf.GetQconf())
}

type AutoMetrics struct {
	Enabled bool
	Qurl    string
	Url     string
	Deurl   string
}

func (am *AutoMetrics) init(qconf *qconf.Qconf) {
	am.Enabled = qconf.GetBool("qms.metrics.autometrics.enabled", getDefaultBoolByEnv("autometrics.enabled"))
	am.Qurl = qconf.GetString("qms.metrics.autometrics.qurl", getDefaultStringByEnv("autometrics.qurl"))
	am.Url = qconf.GetString("qms.metrics.autometrics.url", getDefaultStringByEnv("autometrics.url"))
	am.Deurl = qconf.GetString("qms.metrics.autometrics.delurl", getDefaultStringByEnv("autometrics.delurl"))
}

type Healthy struct {
	PingDisabled bool
	PingPath     string
	HcDisabled   bool
	HcPath       string
}

func (h *Healthy) init() {
	// 兼容老的配置项
	if disabled := qconf.GetBool("qms.healthy.disabled"); disabled { // 后续如果可以通过setXXX设置disabled，则不能直接return
		h.PingDisabled = true
		h.HcDisabled = true
	} else {
		h.PingDisabled = qconf.GetBool("qms.healthy.ping_disabled", qconf.GetBool("qms.healthy.pingDisabled"))
		h.HcDisabled = qconf.GetBool("qms.healthy.hc_disabled", qconf.GetBool("qms.healthy.hcDisabled"))
	}
	h.PingPath = qconf.GetString("qms.healthy.ping_path", qconf.GetString("qms.healthy.pingPath", qconf.GetString("qms.healthy.apiPath", constutil.DefaultPingPath)))
	h.HcPath = qconf.GetString("qms.healthy.hc_path", qconf.GetString("qms.healthy.hcPath", constutil.DefaultHcPath))
}

type Trace struct {
	Disabled bool
	Setting  struct {
		SamplingRate      string `yaml:"sampling_rate"`
		MaxTagValueLength int    `yaml:"max_tag_value_length"`
		TraceFile         string `yaml:"trace_file_name"`
	} `yaml:"settings"`
}

func (t *Trace) init() {
	t.Disabled = qconf.GetBool("qms.tracing.disabled")
	t.Setting.SamplingRate = qconf.GetString("qms.tracing.settings.sampling_rate", qconf.GetString("qms.tracing.settings.samplingRate", getDefaultStringByEnv("tracing.sampling_rate")))
	t.Setting.MaxTagValueLength = qconf.GetInt("qms.tracing.settings.max_tag_value_length", 0)
	t.Setting.TraceFile = qconf.GetString("qms.tracing.settings.trace_file_name", qconf.GetString("qms.tracing.settings.traceFileName", defaultTraceFileName))
}

type AccessLog struct {
	Enabled      bool
	AsyncEnabled bool
	FileName     string
}

func (a *AccessLog) init() {
	a.Enabled = qconf.GetBool("qms.access_log.enabled", qconf.GetBool("qms.accessLog.enabled"))
	a.AsyncEnabled = qconf.GetBool("qms.access_log.async_enabled", qconf.GetBool("qms.accessLog.asyncEnabled", defaultAccessLogAsyncEnabled))
	a.FileName = qconf.GetString("qms.access_log.file_name", qconf.GetString("qms.accessLog.fileName"))
}

type PProf struct {
	Enabled bool
}

func (p *PProf) init() {
	p.Enabled = qconf.GetBool("qms.pprof.enabled", qconf.GetBool("qms.pprof.enabled"))
}

type Native struct {
	Port    int
	address string
}

func (n *Native) init() {
	n.Port = qconf.GetInt("qms.native.port", defaultNativePort)
	n.address = fmt.Sprintf("0.0.0.0:%d", n.Port)
}

func (n *Native) Address() string {
	return n.address
}

// 全局服务治理相关
func initAdvance() {
	// graceful在配置解析之前已经预处理了 此处不需要解析
	conf.Qms.Registry.init(qconf.GetQconf())
	conf.Qms.Metrics.init()
	conf.Qms.Healthy.init()
	conf.Qms.Trace.init()
	conf.Qms.AccessLog.init()
	conf.Qms.PProf.init()
	conf.Qms.Native.init()
}

// 以下代码为了获取大写key，用于兼容过渡，后续可以删掉

var advance = make(map[string]interface{})

func unmarshalAdvance() error {
	fs := afero.NewOsFs()
	exist, err := afero.Exists(fs, fileutil.AdvancedConfigPath())
	if err != nil {
		return errors.WithStack(err)
	}
	if !exist {
		return nil
	}
	buf, err := afero.ReadFile(fs, fileutil.AdvancedConfigPath())
	if err != nil {
		return errors.WithStack(err)
	}
	if err = yaml.Unmarshal(buf, advance); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func searchAdvance(key string) interface{} {
	path := strings.Split(key, ".")
	return searchMapWithPathPrefixes(advance, path)
}

func searchMapWithPathPrefixes(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.Join(path[0:i], ".")

		next, ok := source[prefixKey]
		if ok {
			// Fast path
			if i == len(path) {
				return next
			}

			// Nested case
			var val interface{}
			switch next.(type) {
			case map[interface{}]interface{}:
				val = searchMapWithPathPrefixes(cast.ToStringMap(next), path[i:])
			case map[string]interface{}:
				// Type assertion is safe here since it is only reached
				// if the type of `next` is the same as the type being asserted
				val = searchMapWithPathPrefixes(next.(map[string]interface{}), path[i:])
			default:
				// got a value but nested key expected, do nothing and look for next prefix
			}
			if val != nil {
				return val
			}
		}
	}

	// not found
	return nil
}
