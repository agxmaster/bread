package consul

import (
	"os"
	"strings"
	"time"
)

const (
	DefaultServiceCheckInterval           = "5s"
	MinServiceCheckInterval               = 3 * time.Second
	DefaultServiceDeregisterCriticalAfter = "24h"
	DefaultServiceWatchTimeout            = 5 * time.Minute
	DefaultServiceWatchLatency            = time.Second
	DefaultDegradeThreshold               = 0
	DefaultEmergencyThreshold             = 80
	DefaultCacheSyncInterval              = 3 * time.Hour
	DefaultCalmInterval                   = 1 * time.Hour
	DefaultRetryTimes                     = 3
)

// consul 降级策略
type DegradeStatus int

const (
	WatchNormalize DegradeStatus = iota
	WatchDegraded
)

var DefaultServiceMeta map[string]string

func init() {
	meta := map[string]string{
		"cloud":     "aliyun",
		"container": "vm",
		"registry":  "consul",
	}

	// fill container with POD_IP
	if ipv4 := os.Getenv("POD_IP"); len(ipv4) > 0 {
		meta["container"] = "k8s"
	}

	// file env if exists
	if env := os.Getenv("ENV"); len(env) > 0 {
		meta["env"] = env
	}

	// fill zone with hostname
	hostname, err := os.Hostname()
	if err == nil && len(hostname) > 0 {
		// format in bjf-xxx
		fields := strings.Split(hostname, "-")

		// use the first characters as zone
		if len(fields[0]) > 0 {
			meta["zone"] = fields[0]
		} else {
			meta["zone"] = hostname
		}
	}

	DefaultServiceMeta = meta
}
