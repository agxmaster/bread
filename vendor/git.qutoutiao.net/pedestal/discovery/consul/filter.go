package consul

import (
	"math"
	"sync/atomic"

	"git.qutoutiao.net/pedestal/discovery/logger"
	"git.qutoutiao.net/pedestal/discovery/metrics"
	"github.com/hashicorp/consul/api"
)

// 每个降级都在 watch 的 goroutine 里执行
type Filter interface {
	Apply(entries []*api.ServiceEntry) []*api.ServiceEntry
}

// passingOnlyFilter implements Filter for consul adapter.
type passingOnlyFilter struct {
	key                *watchKey
	totalNodes         int32
	passingOnly        bool
	passingDegrade     bool
	passingThreshold   float64
	emergencyThreshold float64
}

func newPassingOnlyFilter(key *watchKey, opts *option) *passingOnlyFilter {
	pf := &passingOnlyFilter{
		key:                key,
		passingOnly:        opts.passingOnly,
		passingThreshold:   float64(opts.threshold),
		emergencyThreshold: float64(opts.emergency),
	}

	return pf
}

// Apply filters whether service should degrade?
// 	1. return entries if totalEntries == totalPassingEntries
//	2. return passingEntries if totalEntries = totalPassingEntries + totalMaintEntries
// 	3. return entries - maint entries if passingOnly == false
// 	4. return entries - maint entries if emergency triggered
// 	5. return entries - maint entries if degrade triggered
// 	6. return passing entries else
func (pf *passingOnlyFilter) Apply(entries []*api.ServiceEntry) []*api.ServiceEntry {
	if pf == nil {
		return entries
	}

	passingEntries := ReduceConsulHealthWithPassingOnly(entries)
	maintEntries := ReduceConsulHealthWithMaint(entries)

	totalEntries := len(entries)
	totalPassingEntries := len(passingEntries)
	totalMaintEntries := len(maintEntries)

	// check or update totalNodes
	pf.initOrResetTotalNodes(totalEntries, totalMaintEntries)

	switch totalEntries {
	case totalPassingEntries: // all entries are healthy
		logger.Infof("Refresh service(%s): total=%d, passing=%d, maint=%d", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries)

		pf.passingDegrade = false
		metrics.GetMetrics().ReportConsulDegrade(pf.key.name, false)

		return entries

	case totalPassingEntries + totalMaintEntries: // there is no critical entry
		logger.Infof("Refresh service(%s): total=%d, passing=%d, maint=%d", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries)

		pf.passingDegrade = false
		metrics.GetMetrics().ReportConsulDegrade(pf.key.name, false)

		return passingEntries

	}

	// for none passingOnly=true
	if !pf.passingOnly {
		logger.Infof("Refresh service(%s) without passing only: total=%d, passing=%d, maint=%d", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries)

		pf.passingDegrade = false
		metrics.GetMetrics().ReportConsulDegrade(pf.key.name, false)

		return ReduceConsulHealthWithoutMaint(entries)
	}

	// should it degrade with emergency?
	if pf.shouldEmergency(totalEntries, totalPassingEntries, totalMaintEntries) {
		logger.Infof("Degrade service(%s) with for emergency: total=%d, passing=%d, maint=%d, threshold=%.2f", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries, pf.emergencyThreshold)

		pf.passingDegrade = true
		metrics.GetMetrics().ReportConsulDegrade(pf.key.name, true)

		return ReduceConsulHealthWithoutMaint(entries)
	}

	// should it degrade with passing entries only?
	if pf.shouldDegrade(totalPassingEntries) {
		logger.Warnf("Degrade service(%s) with passing only: total=%d, passing=%d, maint=%d, threshold=%.2f", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries, pf.passingThreshold)

		pf.passingDegrade = true
		metrics.GetMetrics().ReportConsulDegrade(pf.key.name, true)

		return ReduceConsulHealthWithoutMaint(entries)
	}

	if pf.passingDegrade {
		logger.Infof("Recover service(%s) with passing only: total=%d, passing=%d, maint=%d, threshold=%.2f", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries, pf.passingThreshold)

		pf.passingDegrade = false
	} else {
		logger.Infof("Refresh service(%s) with passing only: total=%d, passing=%d, maint=%d, threshold=%.2f", pf.key.name, totalEntries, totalPassingEntries, totalMaintEntries, pf.passingThreshold)
	}
	metrics.GetMetrics().ReportConsulDegrade(pf.key.name, false)

	return passingEntries
}

// initOrResetTotalNodes 调整服务 totalNodes 的值
// 	1. totalNodes = totalNodes - totalMaintEntries
func (pf *passingOnlyFilter) initOrResetTotalNodes(totalEntries, totalMaintEntries int) {
	prevTotalEntries := atomic.LoadInt32(&pf.totalNodes)
	nextTotalEntries := int32(totalEntries - totalMaintEntries)

	if atomic.CompareAndSwapInt32(&pf.totalNodes, prevTotalEntries, nextTotalEntries) {
		metrics.GetMetrics().ReportConsulThreshold(pf.key.name, totalEntries)

		logger.Infof("Total nodes of service(%s) changed from %v to %v(total=%d, maint=%d): OK!", pf.key.name, prevTotalEntries, nextTotalEntries, totalEntries, totalMaintEntries)
	} else {
		logger.Warnf("Total nodes of service(%s) changes from %v to %v(total=%d, maint=%d): Failed!", pf.key.name, prevTotalEntries, nextTotalEntries, totalEntries, totalMaintEntries)
	}

	return
}

func (pf *passingOnlyFilter) shouldEmergency(totalEntries, totalPassingEntries, totalMaintEntries int) bool {
	if pf.emergencyThreshold < 0 {
		pf.emergencyThreshold = DefaultEmergencyThreshold
	}

	if pf.emergencyThreshold == 0 {
		return false
	}

	totalCriticalEntries := totalEntries - totalMaintEntries - totalPassingEntries
	if totalCriticalEntries <= 0 {
		return false
	}

	totalValidEntries := totalEntries - totalMaintEntries
	if totalValidEntries <= 0 {
		return false
	}

	return 100*float64(totalCriticalEntries)/float64(totalValidEntries) > pf.emergencyThreshold
}

func (pf *passingOnlyFilter) shouldDegrade(current int) bool {
	if pf.passingThreshold <= 0 {
		return false
	}

	current64 := float64(current)
	totalNodes := float64(atomic.LoadInt32(&pf.totalNodes))
	degradeNodes := math.Round(totalNodes * pf.passingThreshold)

	switch totalNodes {
	case 1, 2:
		if degradeNodes < 1 {
			return current64 < degradeNodes
		}

		return current64 < 1
	case 3, 4:
		if degradeNodes < 2 {
			return current64 < degradeNodes
		}

		return current64 < 2
	case 5, 6:
		if degradeNodes < 3 {
			return current64 < degradeNodes
		}

		return current64 < 3
	case 7, 8:
		if degradeNodes < 4 {
			return current64 < degradeNodes
		}

		return current64 < 4
	case 9, 10:
		if degradeNodes < 5 {
			return current64 < degradeNodes
		}

		return current64 < 5
	default:
		return current64 < degradeNodes
	}
}
