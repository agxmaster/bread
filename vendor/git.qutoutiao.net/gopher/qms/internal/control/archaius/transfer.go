package archaius

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/control"
	"git.qutoutiao.net/gopher/qms/internal/core/client"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/config/model"
	"git.qutoutiao.net/gopher/qms/internal/core/loadbalancer"
	"git.qutoutiao.net/gopher/qms/internal/pkg/backoff"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qms/third_party/forked/afex/hystrix-go/hystrix"
)

//SaveToLBCache save configs
func SaveToLBCache(raw *model.LoadBalancing) {
	qlog.Trace("Loading lb config from archaius into cache")
	oldKeys := LBConfigCache.Items()
	newKeys := make(map[string]bool)
	// if there is no config, none key will be updated
	if raw != nil {
		newKeys = reloadLBCache(raw)
	}
	// remove outdated keys
	for old := range oldKeys {
		if _, ok := newKeys[old]; !ok {
			LBConfigCache.Delete(old)
		}
	}

}
func saveDefaultLB(raw *model.LoadBalancing) string { // return updated key
	c := control.LoadBalancingConfig{
		Enabled:                 raw.Enabled,
		Strategy:                raw.Strategy["name"],
		RetryEnabled:            raw.RetryEnabled,
		RetryOnSame:             raw.RetryOnSame,
		RetryOnNext:             raw.RetryOnNext,
		BackOffKind:             raw.Backoff.Kind,
		BackOffMin:              raw.Backoff.MinMs,
		BackOffMax:              raw.Backoff.MaxMs,
		SessionTimeoutInSeconds: raw.SessionStickinessRule.SessionTimeoutInSeconds,
		SuccessiveFailedTimes:   raw.SessionStickinessRule.SuccessiveFailedTimes,
	}

	setDefaultLBValue(&c)
	LBConfigCache.Set("", c, 0)
	return ""
}
func saveEachLB(k string, raw model.LoadBalancingSpec) string { // return updated key
	c := control.LoadBalancingConfig{
		Enabled:                 true,
		Strategy:                raw.Strategy["name"],
		RetryEnabled:            raw.RetryEnabled,
		RetryOnSame:             raw.RetryOnSame,
		RetryOnNext:             raw.RetryOnNext,
		BackOffKind:             raw.Backoff.Kind,
		BackOffMin:              raw.Backoff.MinMs,
		BackOffMax:              raw.Backoff.MaxMs,
		SessionTimeoutInSeconds: raw.SessionStickinessRule.SessionTimeoutInSeconds,
		SuccessiveFailedTimes:   raw.SessionStickinessRule.SuccessiveFailedTimes,
	}
	setDefaultLBValue(&c)
	LBConfigCache.Set(k, c, 0)
	return k
}

func setDefaultLBValue(c *control.LoadBalancingConfig) {
	if c.Strategy == "" {
		c.Strategy = loadbalancer.StrategyWeightedRoundRobin
	}
	if c.BackOffKind == "" {
		c.BackOffKind = backoff.DefaultBackOffKind
	}
}

//SaveToCBCache save configs
func SaveToCBCache(raw *model.HystrixConfig) {
	qlog.Trace("Loading cb config from archaius into cache")
	oldKeys := CBConfigCache.Items()
	newKeys := make(map[string]bool)
	// if there is no config, none key will be updated
	if raw != nil {
		client.SetTimeoutToClientCache(raw.IsolationProperties)
		newKeys = reloadCBCache(raw)
	}
	// remove outdated keys
	for old := range oldKeys {
		if _, ok := newKeys[old]; !ok {
			CBConfigCache.Delete(old)
		}
	}
}

func saveEachCB(serviceName, serviceType string) string { //return updated key
	command := serviceType
	if serviceName != "" {
		command = strings.Join([]string{serviceType, serviceName}, ".")
	}

	c := hystrix.CommandConfig{}
	if serviceType == common.Consumer { // consumer使用upstream进行替换
		cb := config.GetUpstream(serviceName).CircuitBreaker
		c.ForceFallback = coreconf.GetForceFallback(serviceName, serviceType)
		c.MaxConcurrentRequests = config.GetUpstream(serviceName).Transport.MaxConcurrent // 需要注意这个由transport影响
		c.ErrorPercentThreshold = cb.ErrorThresholdPercentage
		c.RequestVolumeThreshold = cb.RequestVolumeThreshold
		c.SleepWindow = cb.SleepWindowMs
		c.ForceClose = cb.ForceClosed
		c.ForceOpen = cb.ForceOpen
		c.CircuitBreakerEnabled = cb.Enabled
	} else { // provider保持原状
		c.ForceFallback = coreconf.GetForceFallback(serviceName, serviceType)
		c.MaxConcurrentRequests = coreconf.GetMaxConcurrentRequests(command, serviceType)
		c.ErrorPercentThreshold = coreconf.GetErrorPercentThreshold(command, serviceType)
		c.RequestVolumeThreshold = coreconf.GetRequestVolumeThreshold(command, serviceType)
		c.SleepWindow = coreconf.GetSleepWindow(command, serviceType)
		c.ForceClose = coreconf.GetForceClose(command, serviceType)
		c.ForceOpen = coreconf.GetForceOpen(command, serviceType)
		c.CircuitBreakerEnabled = coreconf.GetCircuitBreakerEnabled(command, serviceType)
	}
	cbcCacheKey := GetCBCacheKey(serviceName, serviceType)
	cbcCacheValue, b := CBConfigCache.Get(cbcCacheKey)
	formatString := "save circuit breaker config [%#v] for [%s] "
	if !b || cbcCacheValue == nil {
		if serviceType == "Consumer" {
			qlog.Infof(formatString, c, cbcCacheKey)
		}
		CBConfigCache.Set(cbcCacheKey, c, 0)
		return cbcCacheKey
	}
	commandConfig, ok := cbcCacheValue.(hystrix.CommandConfig)
	if !ok {
		qlog.Infof(formatString, c, serviceName)
		CBConfigCache.Set(cbcCacheKey, c, 0)
		return cbcCacheKey
	}
	if c == commandConfig {
		return cbcCacheKey
	}
	qlog.Infof(formatString, c, serviceName)
	CBConfigCache.Set(cbcCacheKey, c, 0)
	return cbcCacheKey
}

//GetCBCacheKey generate cache key
func GetCBCacheKey(serviceName, serviceType string) string {
	key := serviceType
	if serviceName != "" {
		key = serviceType + ":" + serviceName
	}
	return key
}

func reloadLBCache(src *model.LoadBalancing) map[string]bool { //return updated keys
	keys := make(map[string]bool)
	k := saveDefaultLB(src)
	keys[k] = true
	if src.AnyService == nil {
		return keys
	}
	for name, conf := range src.AnyService {
		k = saveEachLB(name, conf)
		keys[k] = true
	}
	return keys
}

func reloadCBCache(src *model.HystrixConfig) map[string]bool { //return updated keys
	keys := make(map[string]bool)
	// global level config
	k := saveEachCB("", common.Consumer)
	keys[k] = true
	k = saveEachCB("", common.Provider)
	keys[k] = true
	// get all services who have configs
	consumers := make([]string, 0)
	providers := make([]string, 0)
	consumerMap := map[string]bool{}
	providerMap := map[string]bool{}

	// if a service has configurations of IsolationProperties|
	// CircuitBreakerProperties|FallbackPolicyProperties|FallbackProperties,
	// it's configuration should be added to cache when framework starts
	for _, p := range []interface{}{
		src.IsolationProperties,
		src.CircuitBreakerProperties,
		src.FallbackProperties,
		coreconf.GetHystrixConfig().FallbackPolicyProperties} {
		if services, err := getServiceNamesByServiceTypeAndAnyService(p, common.Consumer); err != nil {
			qlog.Errorf("Parse services from config failed: %v", err.Error())
		} else {
			consumers = append(consumers, services...)
		}
		if services, err := getServiceNamesByServiceTypeAndAnyService(p, common.Provider); err != nil {
			qlog.Errorf("Parse services from config failed: %v", err.Error())
		} else {
			providers = append(providers, services...)
		}
	}
	// remove duplicate service names
	for _, name := range consumers {
		consumerMap[name] = true
	}
	for _, name := range providers {
		providerMap[name] = true
	}
	// service level config
	for name := range consumerMap {
		k = saveEachCB(name, common.Consumer)
		keys[k] = true
	}
	for name := range providerMap {
		k = saveEachCB(name, common.Provider)
		keys[k] = true
	}
	return keys
}

func getServiceNamesByServiceTypeAndAnyService(i interface{}, serviceType string) (services []string, err error) {
	// check type
	tmpType := reflect.TypeOf(i)
	if tmpType.Kind() != reflect.Ptr {
		return nil, errors.New("input must be an ptr")
	}
	// check value
	tmpValue := reflect.ValueOf(i)
	if !tmpValue.IsValid() {
		return []string{}, nil
	}

	inType := tmpType.Elem()
	propertyName := inType.Name()

	formatFieldNotExist := "field %s not exist"
	formatFieldNotExpected := "field %s is not type %s"
	// check type
	tmpFieldType, ok := inType.FieldByName(serviceType)
	if !ok {
		return nil, fmt.Errorf(formatFieldNotExist, propertyName+"."+serviceType)
	}
	if tmpFieldType.Type.Kind() != reflect.Ptr {
		return nil, fmt.Errorf(formatFieldNotExpected, propertyName+"."+serviceType, reflect.Ptr)
	}
	// check value
	inValue := reflect.Indirect(tmpValue)
	tmpFieldValue := inValue.FieldByName(serviceType)
	if !tmpFieldValue.IsValid() {
		return []string{}, nil
	}

	anyServiceFieldName := "AnyService"
	//check type
	fieldType := tmpFieldType.Type.Elem()
	tmpAnyServiceFieldType, ok := fieldType.FieldByName(anyServiceFieldName)
	if !ok {
		return nil, fmt.Errorf(formatFieldNotExist, propertyName+"."+serviceType+"."+anyServiceFieldName)
	}
	if tmpAnyServiceFieldType.Type.Kind() != reflect.Map {
		return nil, fmt.Errorf(formatFieldNotExpected, propertyName+"."+serviceType+"."+anyServiceFieldName, reflect.Map)
	}
	// check value
	fieldValue := reflect.Indirect(tmpFieldValue)
	anyServiceFieldValue := fieldValue.FieldByName(anyServiceFieldName)
	if !anyServiceFieldValue.IsValid() {
		return []string{}, nil
	}

	// get service names
	names := anyServiceFieldValue.MapKeys()
	services = make([]string, 0)
	for _, name := range names {
		if name.Kind() != reflect.String {
			return nil, fmt.Errorf(formatFieldNotExpected, "key of "+propertyName+"."+serviceType+"."+anyServiceFieldName, reflect.String)
		}
		services = append(services, name.String())
	}
	return services, nil
}
