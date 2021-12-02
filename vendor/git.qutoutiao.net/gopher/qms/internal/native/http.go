package native

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/pkg/json"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

type Native struct {
	*http.Server
	mux      *http.ServeMux
	routes   []string
	address  string
	listener net.Listener
}

func NewNative() *Native {
	return &Native{
		mux:     http.NewServeMux(),
		address: config.Get().Native.Address(),
		routes:  []string{},
	}
}

func (n *Native) Init(addrM map[string]net.Listener) {
	// 获取全部治理路由
	n.HandleFunc("/routes", func(resp http.ResponseWriter, req *http.Request) {
		json.NewEncoder(resp).Encode(n.routes)
	})

	// ping
	n.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// metrics
	if m := config.Get().Metrics; m.Enabled {
		metricPath := m.Path
		if !strings.HasPrefix(metricPath, "/") {
			metricPath = "/" + metricPath
		}
		qlog.Info("enable metrics API on " + metricPath)
		n.HandleFunc(metricPath, metrics.Handle)
	}

	// pprof
	//if config.Get().PProf.Enabled {
	//	n.HandleFunc("/debug/pprof/", pprof.Index)
	//	n.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//	n.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//	n.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//	n.HandleFunc("/debug/pprof/trace", pprof.Trace)
	//	n.HandleFunc("/debug/pprof/heap", heap)
	//	n.HandleFunc("/debug/pprof/block", block)
	//	n.HandleFunc("/debug/pprof/mutex", mutex)
	//	n.HandleFunc("/debug/pprof/goroutine", goroutine)
	//}
	//
	//// 注册apidoc
	//n.HandleFunc("/apidoc/index.html", apidocIndex)
	//n.HandleFunc("/apidoc/swagger.json", apidocJson)

	if lis, ok := addrM[n.address]; ok {
		n.listener = lis
	}
}

// HandleFunc ...
func (n *Native) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	// todo: 增加安全管控
	n.mux.HandleFunc(pattern, handler)
	n.routes = append(n.routes, pattern)
}

func (n *Native) Run() (err error) {
	n.Server = &http.Server{
		Handler: n.mux,
		Addr:    n.address,
	}

	if n.listener != nil {
		qlog.Infof("native server listening on: %s", n.listener.Addr().String())
		return n.Serve(n.listener)
	}
	qlog.Infof("native server listening on: %s", n.address)
	return n.ListenAndServe()
}

func (n *Native) Stop() error {
	if n.Server == nil {
		qlog.Info("native server never started")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	n.Server.SetKeepAlivesEnabled(false)

	//only golang 1.8 support graceful shutdown.
	if err := n.Server.Shutdown(ctx); err != nil {
		qlog.Warn("http shutdown error: " + err.Error())
		return err // failure/timeout shutting down the server gracefully
	}
	return nil
}
