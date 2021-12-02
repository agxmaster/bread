package degrade

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// degrade 单个服务的降级
type degrade struct {
	cluster                string
	mux                    sync.RWMutex
	status                 degradeStatus
	doUpdateList           UpdateListData     // 用于执行更新通知
	pingCtxCancel          context.CancelFunc // 用于自动退出循环健康检查
	latestAdapterList      []*Node            // Adapter 返回的最新的正常节点列表
	historyEndpoints       *loopList          // *elem 历史节点队列(退出自我保护后会重置)
	stableHistoryEndpoints *loopList          // *elem 历史节点队列(一直按照时间更新)
	applyList              []*Node            // 正在应用的节点列表(Normal 状态时和 latestAdapterList 一致, SelfProtection 状态时是 latestAdapterList+latestAdapterList经过健康检查后的节点)
	degradeClose           chan struct{}      // 用于通知降级开关已经关闭
	//healthList             []*Node            // 通过健康检查的节点
	//isUpstreamDiffer       bool               // 现在使用的数据是否与上游推送产生了便宜
}

func (d *degrade) getHistoryAndUpdate(opts DegradeOpts, latestAdapterList []*Node, doUpdateList UpdateListData) (degradeStatus, *Elem, *Elem) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// 更新 latestAdapterList
	d.latestAdapterList = make([]*Node, len(latestAdapterList))
	for i := range latestAdapterList {
		d.latestAdapterList[i] = latestAdapterList[i]
	}

	// 更新 doUpdateList
	d.doUpdateList = doUpdateList

	return d.status, d.historyEndpoints.back(opts), d.stableHistoryEndpoints.back(opts)
}

func (d *degrade) pushHistoryAndApplyList(opts DegradeOpts, discoveryList []*Node) {
	d.mux.Lock()
	defer d.mux.Unlock()
	// 更新 historyEndpoints
	elnow := &Elem{
		UnixNano: time.Now().UnixNano(),
		List:     discoveryList,
	}
	d.historyEndpoints.push(elnow, opts)
	d.stableHistoryEndpoints.push(elnow, opts)
	d.applyList = d.latestAdapterList // 使用 Adapter 的数据

	//if d.isUpstreamDiffer {
	//	d.isUpstreamDiffer = false
	//	EventSet(EventDegradeUpstreamDifferStatus, d.cluster, MetricsRecover)
	//}
}

func (d *degrade) getStableHistoryEndpoints(opts DegradeOpts) (historyEndpoints, stableHistoryEndpoints *Elem) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	historyEndpoints = d.historyEndpoints.back(opts)
	stableHistoryEndpoints = d.stableHistoryEndpoints.back(opts)
	return
}

func (d *degrade) getLatestAdapterList() (degradeStatus, []*Node) {
	d.mux.RLock()
	defer d.mux.RUnlock()
	return d.status, d.latestAdapterList
}

func (d *degrade) getDegradeStatus() degradeStatus {
	d.mux.RLock()
	defer d.mux.RUnlock()
	return d.status
}

//func (d *degrade) recoverUpstreamDiffer() {
//	d.mux.Lock()
//	defer d.mux.Unlock()
//	if d.isUpstreamDiffer {
//		d.isUpstreamDiffer = false
//		EventSet(EventDegradeUpstreamDifferStatus, d.cluster, MetricsRecover)
//	}
//}

// startSelfProtection 进入自我保护模式
func (d *degrade) startSelfProtection() {
	ctx := func() context.Context {
		if d.getDegradeStatus() != statusNormal {
			return nil
		}

		d.mux.Lock()
		defer d.mux.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		d.status = statusSelfProtection
		d.pingCtxCancel = cancel
		return ctx
	}()

	if ctx == nil {
		return
	}

	Log.Warnf("cluster:%s startSelfProtection", d.cluster)
	go d.pingLoop(ctx)
}

// stopSelfProtection 退出自我保护模式
func (d *degrade) stopSelfProtection(want int) error {
	if d.status == statusNormal {
		return nil
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	if want >= 0 && len(d.latestAdapterList) != want {
		return fmt.Errorf("cluster:%s adapter num not equal want num: %d", d.cluster, want)
	}

	Log.Warnf("cluster:%s stopSelfProtection", d.cluster)

	d.status = statusNormal
	// 取消 ping 循环
	if d.pingCtxCancel != nil {
		d.pingCtxCancel()
		d.pingCtxCancel = nil
	}

	// 清空历史数据，加入当前节点
	d.historyEndpoints.reset()
	opts := GetDegradeOpts()
	// 更新 historyEndpoints
	elnow := &Elem{
		UnixNano: time.Now().UnixNano(),
		List:     d.latestAdapterList,
	}

	// 自我保护自动退出，采用通过健康检查的数据
	//if want == selfProtectionAutoStop {
	//	elnow.List = d.latestAdapterList
	//}

	d.historyEndpoints.push(elnow, opts)
	d.stableHistoryEndpoints.push(elnow, opts)
	d.applyList = d.latestAdapterList // 使用 Adapter 的数据

	// 修改退出后为服务发现的数据, 而不是ping后的数据
	if d.doUpdateList != nil && len(d.latestAdapterList) > 0 {
		d.doUpdateList(d.latestAdapterList, d.status == statusPanic)
	}

	//if !listSame(d.healthList, d.latestAdapterList) {
	//	//EventSet(EventDegradeUpstreamDifferStatus, d.cluster, MetricsDegradeUpstreamDifferStatus)
	//	//d.isUpstreamDiffer = true
	//	healthList, _ := json.Marshal(d.healthList)
	//	latestAdapterList, _ := json.Marshal(d.latestAdapterList)
	//	Log.Warnf("cluster:%s check upstream node list differ. healthList: %v  latestAdapterList: %v", d.cluster, string(healthList), string(latestAdapterList))
	//}

	return nil
}

func (d *degrade) pingLoop(ctx context.Context) {
	opts := GetDegradeOpts()
	//healthCheckVersion := uint64(0)
	historyEndpoints, stableHistoryEndpoints := d.getStableHistoryEndpoints(opts)

	historyEndpointsStr, _ := json.Marshal(historyEndpoints)
	stableHistoryEndpointsStr, _ := json.Marshal(stableHistoryEndpoints)
	Log.Warnf("cluster:%s start pingLoop historyEndpoints:%s stableHistoryEndpoints: %s", d.cluster, string(historyEndpointsStr), string(stableHistoryEndpointsStr))

	st := time.Now()
	updateHistoryList := make([]*Node, 0)
	ti := time.NewTimer(opts.HealthCheckInterval)
	defer func() {
		ti.Stop()
		EventDeleteLabelValues(EeventDegradeSelfProtectionTime, d.cluster)
		EventDeleteLabelValues(EventDegradePanicStatus, d.cluster)
	}()

	if !d.continueDoPing(ctx, historyEndpoints.List, stableHistoryEndpoints.List, &updateHistoryList) {
		Log.Warnf("cluster:%s stop pingLoop", d.cluster)
		return
	}

	EventSet(EeventDegradeSelfProtectionTime, d.cluster, time.Now().Sub(st).Seconds())

	for {
		select {
		case <-ctx.Done():
			return
		case <-ti.C:
			opts = GetDegradeOpts()
			//healthCheckVersion++ // 每次 ping 都是更新 checkchange 的缓存
			// 降级已经被关闭
			if opts.Flag == degradeFlagClose {
				if err := d.stopSelfProtection(-1); err != nil {
					Log.Warnf("cluster:%s stop pingLoop stopSelfProtection error: %v", d.cluster, err)
				} else {
					Log.Warnf("cluster:%s stop pingLoop", d.cluster)
				}
				return
			}

			// 自我保护状态超过最长持续时间，自动退出
			if time.Now().Sub(st) > opts.SelfProtectionMaxTime {
				if err := d.stopSelfProtection(selfProtectionAutoStop); err != nil {
					Log.Warnf("cluster:%s stop pingLoop stopSelfProtection error: %v", d.cluster, err)
				} else {
					Log.Warnf("cluster:%s stop pingLoop", d.cluster)
				}
				return
			}

			// 刷新自我保护进入时间
			EventSet(EeventDegradeSelfProtectionTime, d.cluster, time.Now().Sub(st).Seconds())

			if !d.continueDoPing(ctx, historyEndpoints.List, stableHistoryEndpoints.List, &updateHistoryList) {
				Log.Warnf("cluster:%s stop pingLoop", d.cluster)
				return
			}

			ti.Reset(GetDegradeOpts().HealthCheckInterval)
		case <-d.degradeClose:
			if err := d.stopSelfProtection(-1); err != nil {
				Log.Warnf("cluster:%s stop pingLoop stopSelfProtection error: %v", d.cluster, err)
			} else {
				Log.Warnf("cluster:%s stop pingLoop", d.cluster)
			}

			return
		}
	}
}

func (d *degrade) continueDoPing(ctx context.Context, historyEndpoints, stableHistoryEndpoints []*Node, updateHistoryList *[]*Node) (doPing bool) {
	opts := GetDegradeOpts()
	status, latestAdapterList := d.getLatestAdapterList()
	panicStateChange := false // 由于要向下游传递节点变更和状态变更，这个变量确保在数据没有变更但是状态变更时，可以向下传递

	defer func() {
		if !doPing {
			// 不需要继续 ping, 在 latestAdapterList 没有变化时退出自我保护模式
			if err := d.stopSelfProtection(len(latestAdapterList)); err != nil {
				Log.Errorf("cluster:%s stopSelfProtection err: %v", d.cluster, err)
				doPing = true
			}
		}

		if GetDegradeOpts().Flag == degradeFlagClose {
			return
		}

		// 当数据发生变化，执行更新通知
		d.mux.RLock()
		defer d.mux.RUnlock()
		if panicStateChange || !listSame(*updateHistoryList, d.applyList) {
			if d.doUpdateList != nil {
				d.doUpdateList(d.applyList, d.status == statusPanic)
			}

			*updateHistoryList = d.applyList
		}
	}()

	// 聚合历史数据和当前最新的 Adapter 节点
	endpointds := make([]*Node, 0)
	used := map[string]bool{}
	for i := range latestAdapterList {
		if !used[latestAdapterList[i].GetSelfProtectionID()] {
			used[latestAdapterList[i].GetSelfProtectionID()] = true
			endpointds = append(endpointds, latestAdapterList[i])
		}
	}

	for i := range historyEndpoints {
		if !used[historyEndpoints[i].GetSelfProtectionID()] {
			used[historyEndpoints[i].GetSelfProtectionID()] = true
			endpointds = append(endpointds, historyEndpoints[i])
		}
	}

	// 开始检查,对比检查检查后的数据是否和 latestAdapterList 一致
	tmps, _ := json.Marshal(latestAdapterList)
	Log.Infof("cluster:%s degrade start ping... latestAdapterList(%d): %s", d.cluster, len(latestAdapterList), string(tmps))
	tmps, _ = json.Marshal(endpointds)
	Log.Infof("cluster:%s degrade start ping... endpointds(%d): %s", d.cluster, len(endpointds), string(tmps))

	defer func(now time.Time) {
		Log.Infof("cluster:%s degrade end ping cost:%s", d.cluster, time.Now().Sub(now))
	}(time.Now())

	healthList := make([]*Node, 0)
	for i := range endpointds {
		if err := checkHealth(endpointds[i], opts.PingTimeout); err != nil {
			b, _ := json.Marshal(endpointds[i])
			Log.Warnf("cluster:%s endpoint:%s check health err:%v", endpointds[i].Name, string(b), err)
		} else {
			healthList = append(healthList, endpointds[i])
		}
	}

	//func() {
	//	d.mux.Lock()
	//	defer d.mux.Unlock()
	//	d.healthList = healthList
	//}()

	// 进入或者退出恐慌状态
	switch status {
	case statusSelfProtection:
		// 自我保护下，hc 后的节点数小于恐慌阈值,则进入恐慌保护,使用15分钟前的全量列表
		if len(stableHistoryEndpoints) > 0 && float64(len(healthList))/float64(len(stableHistoryEndpoints)) < opts.PanicThreshold {
			if ctxIsDone(ctx) {
				return true
			}

			func() {
				d.mux.Lock()
				defer d.mux.Unlock()
				if GetDegradeOpts().Flag == degradeFlagOpen && d.status == statusSelfProtection {
					d.status = statusPanic
					d.applyList = stableHistoryEndpoints // 使用绝对的 15 分钟前列表
				}
			}()

			panicStateChange = true
			EventSet(EventDegradePanicStatus, d.cluster, MetricsDegradePanicStatus)
			Log.Warnf("cluster:%s healthList(%d) historyEndpoints(%d) into panic status PanicThreshold:%f",
				d.cluster, len(healthList), len(historyEndpoints), opts.PanicThreshold)
			return true
		}
	case statusPanic:
		// 恐慌状态下，hc 后的节点数/绝对15m前节点 大于自我保护阈值,则退回到自我保护状态,否则继续保持状态
		if len(stableHistoryEndpoints) > 0 && float64(len(healthList))/float64(len(stableHistoryEndpoints)) < opts.Threshold {
			return true
		} else {
			// 恢复到自我保护状态, 继续走自我保护退出或者更新列表逻辑
			if ctxIsDone(ctx) {
				return true
			}

			func() {
				d.mux.Lock()
				defer d.mux.Unlock()
				if d.status == statusPanic {
					d.status = statusSelfProtection
				}
				status = d.status
			}()

			panicStateChange = true
			EventDeleteLabelValues(EventDegradePanicStatus, d.cluster)
			Log.Warnf("cluster:%s healthList(%d) stableHistoryEndpoints(%d) status into selfProtection Threshold:%f",
				d.cluster, len(healthList), len(stableHistoryEndpoints), opts.Threshold)
		}
	}

	// 健康检查发现和 Adapter 不一致, 使用健康检查后的列表
	if GetDegradeOpts().Flag == degradeFlagOpen && d.getDegradeStatus() == statusSelfProtection {
		func() {
			d.mux.Lock()
			defer d.mux.Unlock()
			d.applyList = healthList
		}()
	}

	// 监控检查后的历史节点+ Adapter 列表和 Adapter 节点一致则退出自我保护到正常模式
	if status == statusSelfProtection && listSame(healthList, latestAdapterList) {
		Log.Warnf("cluster:%s healthCheck list same", d.cluster)
		return false
	}

	return true
}
