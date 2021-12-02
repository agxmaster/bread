package invoker

import (
	"strings"
	"sync"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

var rChecker routeChecker

type routeChecker struct {
	mutex sync.Mutex
	cache map[string]bool //<serviceName, ifNoDiscovery>
}

func (rt *routeChecker) isNoDiscovery(service, endpoint, routeType string) bool {
	switch routeType {
	case common.RouteDiscovery:
		return false
	case common.RouteDirect:
		return true
	case common.RouteSidecar:
		if config.GetUpstream(service).Sidecar.Enabled {
			return true
		}
	default:
	}

	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	if rt.cache == nil {
		rt.cache = make(map[string]bool)
	}
	if noDiscovery, exist := rt.cache[endpoint]; exist {
		return noDiscovery
	}
	direct := guessByName(endpoint)
	rt.cache[endpoint] = direct
	if direct {
		qlog.Infof("guess service(%s) access-type: direct", endpoint)
	} else {
		qlog.Infof("guess service(%s) access-type: service-discovery", endpoint)
	}
	return direct
}

func guessByName(serviceName string) bool {
	if strings.Contains(serviceName, ".") {
		return true //normal
	}
	return false //service discovery
}
