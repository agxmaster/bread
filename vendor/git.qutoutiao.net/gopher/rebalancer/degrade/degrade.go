package degrade

import (
	"encoding/json"
	"fmt"
)

var kernelControl *kernel // 所有服务的降级核心控制信息

func init() {
	kernelControl = new(kernel)
	kernelControl.UpdateDegradeOpts(DegradeOpts{})
}

// UpdateDegradeOpts 动态更新服务降级配置
func UpdateDegradeOpts(opts DegradeOpts) {
	optsOld := GetDegradeOpts()
	if opts.Flag == 0 {
		opts.Flag = optsOld.Flag
	}
	if opts.Threshold == 0 {
		opts.Threshold = optsOld.Threshold
	}
	if opts.PanicThreshold == 0 {
		opts.PanicThreshold = optsOld.PanicThreshold
	}
	if opts.ThresholdContrastInterval == 0 {
		opts.ThresholdContrastInterval = optsOld.ThresholdContrastInterval
	}
	if opts.EndpointsSaveInterval == 0 {
		opts.EndpointsSaveInterval = optsOld.EndpointsSaveInterval
	}
	if opts.HealthCheckInterval == 0 {
		opts.HealthCheckInterval = optsOld.HealthCheckInterval
	}
	if opts.PingTimeout == 0 {
		opts.PingTimeout = optsOld.PingTimeout
	}
	if opts.SelfProtectionMaxTime == 0 {
		opts.SelfProtectionMaxTime = optsOld.SelfProtectionMaxTime
	}

	kernelControl.UpdateDegradeOpts(opts)
}

// ResetDegradeOpts 将服务降级配置 恢复到默认值（opts 为零值的部分）
func ResetDegradeOpts(opts DegradeOpts) {
	kernelControl.UpdateDegradeOpts(opts)
}

// GetDegradeOpts 获取服务降级配置
func GetDegradeOpts() DegradeOpts {
	return kernelControl.GetDegradeOpts()
}

// CloseDegrade 关闭服务降级
func CloseDegrade() {
	opts := GetDegradeOpts()
	opts.Flag = degradeFlagClose
	kernelControl.UpdateDegradeOpts(opts)
}

// onUpdateList 判断是否进行降级处理
func UpdateList(serviceName string, discoveryList []*Node, doUpdateList UpdateListData) error {
	if len(serviceName) <= 0 {
		// 没有服务名不走降级，需要在外层传递数据
		return fmt.Errorf("serviceName is null，updateList did not do it")
	}
	degrade := kernelControl.getDegrade(serviceName)

	// 有标记下线的节点直接删除
	discoveryList, offlinesList := getNormalServices(discoveryList)

	opts := GetDegradeOpts()
	// 获取 15 分钟前的节点时，会更新 latestAdapterList 和 doUpdateList
	status, history, stableHistory := degrade.getHistoryAndUpdate(opts, discoveryList, doUpdateList)

	// 非正常模式只更新 latestAdapterList
	if opts.Flag == degradeFlagOpen && status != statusNormal {
		// 上游推送数据后，立即向下游推送数据
		degrade.mux.RLock()
		defer degrade.mux.RUnlock()
		if degrade.doUpdateList != nil && len(degrade.applyList) > 0 {
			degrade.doUpdateList(degrade.applyList, degrade.status == statusPanic)
		}
		return nil
	}

	// 判断是否进入自我保护模式
	if opts.Flag == degradeFlagOpen && history != nil { // 如果第一次获取时注册中心故障了，会使用初始化传入的节点
		// 历史节点中未被标记下线的节点数
		historyCount := 0
		for _, srv := range history.List {
			if _, ok := offlinesList[srv.Address]; !ok {
				historyCount++
			}
		}
		stableHistoryCount := 0
		for _, srv := range stableHistory.List {
			if _, ok := offlinesList[srv.Address]; !ok {
				stableHistoryCount++

			}
		}

		onlineCount := len(discoveryList)
		if onlineCount > 0 {
			onlineCount += addEndpoints
		}

		threshold := float64(onlineCount) / float64(historyCount)
		panicThreshold := float64(len(discoveryList)) / float64(stableHistoryCount)
		if threshold < opts.Threshold || panicThreshold < opts.PanicThreshold {
			// 进入自我保护模式, 变更状态，并开始健康检查
			degrade.startSelfProtection()
			b, _ := json.Marshal(discoveryList)
			Log.Warnf("cluster:%s [adapter consul]degrade can not update list: %s cause:proportion: %f < %f(threshold) or %f < %f(panicThreshold)",
				degrade.cluster, string(b), threshold, opts.Threshold, panicThreshold, opts.PanicThreshold)
			return nil
		}
	}

	// 更新 historyEndpoints 和 applyList 并恢复 UpstreamDiffer
	degrade.pushHistoryAndApplyList(opts, discoveryList)

	//// 退出自我保护模式,并更新 applyList
	//if err := degrade.stopSelfProtection(-1); err != nil {
	//	degrade.log.Warnf("cluster:%s [adapter consul]degrade stopSelfProtection: %v", degrade.cluster, err)
	//}

	// 执行更新通知
	if doUpdateList != nil {
		doUpdateList(discoveryList, false)
	}
	return nil
}
