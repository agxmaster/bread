package handler

import (
	"strconv"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	coreconf "git.qutoutiao.net/gopher/qms/internal/core/config"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	tree "git.qutoutiao.net/gopher/qms/internal/pkg/routetree"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

const (
	maxLength = 100
)

type MetricsProviderHandler struct{}

func (m *MetricsProviderHandler) Name() string {
	return MetricsProvider
}

func newMetricsProviderHandler() Handler {
	return &MetricsProviderHandler{}
}

func (m *MetricsProviderHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !coreconf.GlobalDefinition.Qms.Metrics.Enabled {
		chain.Next(i, cb)
		return
	}

	var path, histogramName, counterName string
	labels := make(map[string]string, 4)
	labels[metrics.ReqProtocolLable] = i.Protocol.String()
	labels[metrics.QMSLabel] = "true"
	if i.Protocol == protocol.ProtocHTTP {
		path = i.BriefURI
		labels[metrics.RespUriLable] = path
		histogramName = metrics.ReqDuration
		counterName = metrics.ReqQPS
	} else {
		path = i.OperationID
		labels[metrics.RespHandlerLable] = path
		histogramName = metrics.GrpcReqDuration
		counterName = metrics.GrpcReqQPS
	}

	st := time.Now()
	chain.Next(i, func(r *invocation.Response) (err error) {
		err = cb(r)

		labels[metrics.RespCodeLable] = strconv.Itoa(r.Status)
		for _, l := range metrics.GetRestCustomLabels() {
			if r.KVs != nil {
				if v, ok := r.KVs[l.LabelValueKey].(string); ok {
					labels[l.LabelName] = v
				} else {
					labels[l.LabelName] = ""
				}
			}
		}

		if err := metrics.HistogramObserve(histogramName, time.Since(st).Seconds(), labels); err != nil {
			qlog.Errorf("HistogramObserve, path: %s status: %d err: %s", path, r.Status, err.Error())
		}

		if err := metrics.CounterAdd(counterName, 1, labels); err != nil {
			qlog.Errorf("CounterAdd, path: %s status: %d err: %s", path, r.Status, err.Error())
		}

		return
	})
}

//consumer
type MetricsConsumerHandler struct {
	routeM       map[string]struct{} // 记录所有route
	serviceMtree map[string]*tree.Node
	mu           sync.RWMutex
}

func (m *MetricsConsumerHandler) Name() string {
	return MetricsConsumer
}

func newMetricsConsumerHandler() Handler {
	serviceMtree := make(map[string]*tree.Node)
	for service, upstream := range config.Get().Upstreams {
		if service == constutil.Common {
			continue
		}
		for _, path := range upstream.Parampath {
			root, ok := serviceMtree[service]
			if !ok {
				root = tree.NewNode()
				serviceMtree[service] = root
			}

			root.AddRoute(path)
		}
	}

	return &MetricsConsumerHandler{
		routeM:       make(map[string]struct{}, maxLength), // 收敛后的MAP
		serviceMtree: serviceMtree,
	}
}

func (m *MetricsConsumerHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	if !coreconf.GlobalDefinition.Qms.Metrics.Enabled {
		chain.Next(i, cb)
		return
	}

	// make labels
	var path, histogramName, counterName string
	labels := map[string]string{
		metrics.RemoteLable:      i.MicroServiceName,
		metrics.ReqProtocolLable: i.Protocol.String(),
	}
	if i.Protocol == protocol.ProtocHTTP {
		if root := m.serviceMtree[i.MicroServiceName]; root != nil {
			path = root.GetFullPath(i.URLPathFormat)
		}
		if path == "" {
			path = i.URLPathFormat
		}
		labels[metrics.RespUriLable] = path
		histogramName = metrics.ClientReqDuration
		counterName = metrics.ClientReqQPS
	} else {
		path = i.OperationID
		labels[metrics.RespHandlerLable] = path
		histogramName = metrics.ClientGrpcReqDuration
		counterName = metrics.ClientGrpcReqQPS
	}

	st := time.Now()
	chain.Next(i, func(r *invocation.Response) error {
		latency := time.Since(st).Seconds()
		labels[metrics.RespCodeLable] = strconv.Itoa(r.Status)
		if !m.isCollect(i.MicroServiceName, path, i.Protocol) {
			return cb(r)
		}

		if err := metrics.HistogramObserve(histogramName, latency, labels); err != nil {
			qlog.Errorf("HistogramObserve, path: %s status: %d err: %s", path, r.Status, err.Error())
		}

		if err := metrics.CounterAdd(counterName, 1, labels); err != nil {
			qlog.Errorf("HistogramObserve, path: %s status: %d err: %s", path, r.Status, err.Error())
		}
		return cb(r)
	})
}

// 是否收集数据 grpc默认收集
func (m *MetricsConsumerHandler) isCollect(sname, path string, proto protocol.Protocol) bool {
	if proto == protocol.ProtocHTTP {
		key := m.getKey(sname, path)

		// 判断是否存在
		m.mu.RLock()
		if _, ok := m.routeM[key]; ok { // 存在
			m.mu.RUnlock()
			return true
		}
		m.mu.RUnlock()

		// 加写锁 第二次判断是否存在
		m.mu.Lock()
		if _, ok := m.routeM[key]; ok { // 存在
			m.mu.Unlock()
			return true
		}
		// 不存在
		if len(m.routeM) >= maxLength { // 如果路由长度超过最大值，则丢弃该路由的metrics收集
			m.mu.Unlock()
			return false
		}
		m.routeM[key] = struct{}{}
		m.mu.Unlock()
	}
	return true
}

func (m *MetricsConsumerHandler) getKey(sname, uri string) string {
	return sname + "-" + uri
}
