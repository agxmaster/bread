package config

const (
	// 默认值
	defaultSidecarEnabled                  = false
	defaultSidecarAddress                  = "127.0.0.1:8102"
	defaultTimeoutMs                       = 1000
	defaultLoadbalance                     = "WeightedRoundRobin"
	defaultMaxConcurrent                   = 5000
	defaultMaxIdleConn                     = 100
	defaultRetryCondition                  = "http_500,http_502,http_503,timeout"
	defaultBackoffKind                     = "zero"
	defaultMaxQPS                          = 2147483647
	defaultCircuitScope                    = "instance"
	defaultCircuitSleepWindowMs            = 15000
	defaultCircuitRequestVolumeThreshold   = 50
	defaultCircuitErrorThresholdPercentage = 50
	defaultEnv                             = "dev"
	defaultVersion                         = "0.0.1"
	defaultRegistryCacheDir                = "/data/app"
	defaultTraceFileName                   = "/data/logs/trace/trace.log"
	defaultAccessLogAsyncEnabled           = true
	//defaultAccessLogFileName               = "/data/logs/app/%s_access.log"
	defaultNativePort = 9080
	// 区分环境的默认值
	defaultRegistryAddressQa         = "http://registry-qa.qutoutiao.net"
	defaultRegistryAddressPrd        = "http://127.0.0.1:8500"
	defaultRegistryAddressK8sPrd     = "http://registry-k8s.qutoutiao.net:8500"
	defaultTraceSamplingRateQa       = "1.0"
	defaultTraceSamplingRatePrd      = "0.01"
	defaultAutometricsEnabled        = false
	defaultAutometricsEnabledQaPrd   = true
	defaultAutometricsQurlQa         = "http://172.25.128.4:8090/api/v1/query"
	defaultAutometricsQurlPrd        = "http://consul-api.qutoutiao.net/api/v1/query"
	defaultAutometricsUrlQa          = "http://172.25.128.4:8090/api/v1/register"
	defaultAutometricsUrlPrd         = "http://consul-api.qutoutiao.net/api/v1/register"
	defaultAutometricsDelurlQa       = "http://172.25.128.4:8090/api/v1/deleteipport"
	defaultAutometricsDelurlPrd      = "http://consul-api.qutoutiao.net/api/v1/deleteipport"
	defaultLoadBalanceStrategy       = "WeightedRoundRobin"
	defaultLoadBalanceWithoutBreaker = false

	Limit = "limit"
)
