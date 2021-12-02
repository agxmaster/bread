package config

import (
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/runtime"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
)

// GetRegistratorDisable returns the Disable of service registry
func GetRegistratorDisable() bool {
	return Get().Registry.Disabled
}

func GetMetricsEnabled() bool {
	return Get().Metrics.AutoMetrics.Enabled
}

//func getDefaultAccessLogFileName() string {
//	return fmt.Sprintf(defaultAccessLogFileName, funcutil.ConvertFileName(Get().Service.AppID))
//}

func getServiceKey(service string, suffix string) string {
	return getkey("qms.service", service, suffix)
}

func getKey(prefix, service, suffix string) string {
	keys := make([]string, 0, 3)
	keys = append(keys, prefix)
	if service != constutil.Common {
		keys = append(keys, service)
	}
	keys = append(keys, suffix)
	return strings.Join(keys, ".")
}

func getkey(keys ...string) string {
	return strings.Join(keys, ".")
}

func getProtocolByName(name string) protocol.Protocol {
	if name == "" {
		return protocol.ProtocUnknown
	}
	names := strings.SplitN(name, "-", 2)
	return protocol.ToProtocol(names[0])
}

func getDefaultStringByEnv(key string) string {
	switch key {
	case "registry.address":
		if isOnline() {
			if runtime.InsideDocker {
				return defaultRegistryAddressK8sPrd
			}
			return defaultRegistryAddressPrd
		}
		return defaultRegistryAddressQa
	case "tracing.sampling_rate":
		if isOnline() {
			return defaultTraceSamplingRatePrd
		}
		return defaultTraceSamplingRateQa
	case "autometrics.qurl":
		if isOnline() {
			return defaultAutometricsQurlPrd
		}
		return defaultAutometricsQurlQa
	case "autometrics.url":
		if isOnline() {
			return defaultAutometricsUrlPrd
		}
		return defaultAutometricsUrlQa
	case "autometrics.delurl":
		if isOnline() {
			return defaultAutometricsDelurlPrd
		}
		return defaultAutometricsDelurlQa
	}
	return ""
}

func getDefaultBoolByEnv(key string) bool {
	switch key {
	case "autometrics.enabled":
		if env := qenv.Get(); env.IsQa() || env.IsPrd() {
			return defaultAutometricsEnabledQaPrd
		}
		return defaultAutometricsEnabled
	}
	return false
}

func isOnline() bool {
	if env := qenv.Get(); env.IsPre() || env.IsPrd() {
		return true
	}
	return false
}
