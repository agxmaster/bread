package config

import (
	"strconv"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
)

// constant for default values of abort and delay
const (
	DefaultAbortPercent = 0
	DefaultAbortStatus  = 0
	DefaultDelayPercent = 0
)

// GetAbortPercent get abort percentage
func GetAbortPercent(protocol, microServiceName, schema, operation string) int {

	var key string
	var abortPercent int
	if microServiceName != "" && schema != "" && operation != "" {
		key = GetFaultInjectionOperationKey(microServiceName, schema, operation)
		abortPercent = qconf.GetInt(GetFaultAbortPercentKey(key, protocol), DefaultAbortPercent)
	}
	if abortPercent == 0 && microServiceName != "" && schema != "" {
		key = GetFaultInjectionSchemaKey(microServiceName, schema)
		abortPercent = qconf.GetInt(GetFaultAbortPercentKey(key, protocol), DefaultAbortPercent)
	}
	if abortPercent == 0 && microServiceName != "" {
		key = GetFaultInjectionServiceKey(microServiceName)
		abortPercent = qconf.GetInt(GetFaultAbortPercentKey(key, protocol), DefaultAbortPercent)
	}
	if abortPercent == 0 {
		key = GetFaultInjectionGlobalKey()
		abortPercent = qconf.GetInt(GetFaultAbortPercentKey(key, protocol), DefaultAbortPercent)
	}

	return abortPercent
}

// GetAbortStatus get abort status
func GetAbortStatus(protocol, microServiceName, schema, operation string) int {

	var key string
	var abortHTTPStatus int
	if microServiceName != "" && schema != "" && operation != "" {
		key = GetFaultInjectionOperationKey(microServiceName, schema, operation)
		abortHTTPStatus = qconf.GetInt(GetFaultAbortHTTPStatusKey(key, protocol), DefaultAbortStatus)
	}
	if abortHTTPStatus == 0 && microServiceName != "" && schema != "" {
		key = GetFaultInjectionSchemaKey(microServiceName, schema)
		abortHTTPStatus = qconf.GetInt(GetFaultAbortHTTPStatusKey(key, protocol), DefaultAbortStatus)
	}
	if abortHTTPStatus == 0 && microServiceName != "" {
		key = GetFaultInjectionServiceKey(microServiceName)
		abortHTTPStatus = qconf.GetInt(GetFaultAbortHTTPStatusKey(key, protocol), DefaultAbortStatus)
	}
	if abortHTTPStatus == 0 {
		key = GetFaultInjectionGlobalKey()
		abortHTTPStatus = qconf.GetInt(GetFaultAbortHTTPStatusKey(key, protocol), DefaultAbortStatus)
	}

	return abortHTTPStatus
}

// GetDelayPercent get delay percentage
func GetDelayPercent(protocol, microServiceName, schema, operation string) int {

	var key string
	var delayPercent int
	if microServiceName != "" && schema != "" && operation != "" {
		key = GetFaultInjectionOperationKey(microServiceName, schema, operation)
		delayPercent = qconf.GetInt(GetFaultDelayPercentKey(key, protocol), DefaultDelayPercent)
	}
	if delayPercent == 0 && microServiceName != "" && schema != "" {
		key = GetFaultInjectionSchemaKey(microServiceName, schema)
		delayPercent = qconf.GetInt(GetFaultDelayPercentKey(key, protocol), DefaultDelayPercent)
	}
	if delayPercent == 0 && microServiceName != "" {
		key = GetFaultInjectionServiceKey(microServiceName)
		delayPercent = qconf.GetInt(GetFaultDelayPercentKey(key, protocol), DefaultDelayPercent)
	}
	if delayPercent == 0 {
		key = GetFaultInjectionGlobalKey()
		delayPercent = qconf.GetInt(GetFaultDelayPercentKey(key, protocol), DefaultDelayPercent)
	}

	return delayPercent
}

// GetFixedDelay get fixed delay
func GetFixedDelay(protocol, microServiceName, schema, operation string) time.Duration {

	var key string
	var fixedDelayTime time.Duration
	var fixedDelay interface{}
	if microServiceName != "" && schema != "" && operation != "" {
		key = GetFaultInjectionOperationKey(microServiceName, schema, operation)
		fixedDelay = qconf.Get(GetFaultFixedDelayKey(key, protocol))
	}
	if fixedDelay == nil && microServiceName != "" && schema != "" {
		key = GetFaultInjectionSchemaKey(microServiceName, schema)
		fixedDelay = qconf.Get(GetFaultFixedDelayKey(key, protocol))
	}
	if fixedDelay == nil && microServiceName != "" {
		key = GetFaultInjectionServiceKey(microServiceName)
		fixedDelay = qconf.Get(GetFaultFixedDelayKey(key, protocol))
	}
	if fixedDelay == nil {
		key = GetFaultInjectionGlobalKey()
		fixedDelay = qconf.Get(GetFaultFixedDelayKey(key, protocol))
	}
	switch fixedDelay.(type) {
	case int:
		fixedDelayInt := fixedDelay.(int)
		fixedDelayTime = time.Duration(fixedDelayInt) * time.Millisecond
	case string:
		fixedDelayInt, _ := strconv.Atoi(fixedDelay.(string))
		fixedDelayTime = time.Duration(fixedDelayInt) * time.Millisecond
	}
	return fixedDelayTime
}
