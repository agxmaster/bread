package ginhttp

import (
	"context"
	"fmt"
	"net/http"
	rt "runtime"
	"strings"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/requestid"
	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"git.qutoutiao.net/gopher/qms/internal/pkg/metrics"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/httputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/gorm"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qms/pkg/redis"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// constants for metric path and name
const (
	//Name is a variable of type string which indicates the protocol being used
	Name = "rest"

	MimeFile = "application/octet-stream"
	MimeMult = "multipart/form-data"
)

func init() {
	server.InstallPlugin(Name, newGinServer)
}

type ginServer struct {
	microServiceName string
	gs               *gin.Engine
	opts             *initOption
	mux              sync.RWMutex
	exit             chan chan error
	server           *http.Server
}

func newGinServer(option *server.InitOptions) server.Server {
	opts := newInitOption(option)
	if !qenv.Get().IsDev() || qlog.GetLevel() != qlog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}
	gs := gin.New()
	gs.Use(wrapHandlerChain(opts))

	// metrics 兼容以前的
	if m := config.Get().Metrics; m.Enabled {
		metricPath := m.Path
		if !strings.HasPrefix(metricPath, "/") {
			metricPath = "/" + metricPath
		}
		qlog.Info("enable metrics API on " + metricPath)
		gs.GET(metricPath, metrics.GinHandleFunc)
	}

	if healthy := config.Get().Healthy; !healthy.PingDisabled {
		pingPath := healthy.PingPath
		if !strings.HasPrefix(pingPath, "/") {
			pingPath = "/" + pingPath
		}
		qlog.Info("enable ping API on " + pingPath)
		pingFunc := func(c *gin.Context) {
			if checkService, appID := c.GetHeader(constutil.ServiceHeader), config.Get().Service.AppID; checkService != "" && checkService != appID {
				info := fmt.Sprintf("qms checkService:%s is not localService:%s", checkService, appID)
				c.String(http.StatusPreconditionFailed, info)
				return
			}
			c.String(http.StatusOK, http.StatusText(http.StatusOK))
		}
		gs.GET(pingPath, pingFunc)
		gs.HEAD(pingPath, pingFunc)
	}

	if config.Get().PProf.Enabled {
		//add pprof
		gs.GET("/debug/pprof/", ginIndex)
		gs.GET("/debug/pprof/cmdline", ginCmdline)
		gs.GET("/debug/pprof/profile", ginProfile)
		gs.GET("/debug/pprof/symbol", ginSymbol)
		gs.GET("/debug/pprof/trace", ginTrace)
		gs.GET("/debug/pprof/heap", ginHeap)
		gs.GET("/debug/pprof/block", ginBlock)
		gs.GET("/debug/pprof/mutex", ginMutex)
		gs.GET("/debug/pprof/goroutine", ginGoroutine)
	}

	bindHc(gs, opts)

	// swagger api
	gs.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	gs.GET("/apidoc/*any", apidoc)

	return &ginServer{
		opts: opts,
		gs:   gs,
	}
}

//wrapHandlerChain wrap business handler with handler chain
func wrapHandlerChain(opts *initOption) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var stacktrace string
				for i := 1; ; i++ {
					_, f, l, got := rt.Caller(i)
					if !got {
						break
					}

					stacktrace += fmt.Sprintf("%s:%d\n", f, l)
				}
				qlog.WithFields(qlog.Fields{
					"path":  ctx.Request.URL.Path,
					"panic": r,
					"stack": stacktrace,
				}).Error("handle request panic.")
				ctx.String(http.StatusInternalServerError, "server got a panic, plz check log.")
			}
		}()

		c, err := handler.GetChain(common.Provider, opts.chainName)
		if err != nil {
			qlog.WithError(err).Error("handler chain init err.")
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		inv, err := HTTPRequest2Invocation(ctx, opts.serverName, "rest", ctx.Request.URL.Path)
		if err != nil {
			qlog.WithError(err).Error("transfer http request to invocation failed.")
			return
		}
		//give inv.Ctx to user handlers, modules may inject headers in handler chain
		c.Next(inv, func(ir *invocation.Response) error {
			if ir.Err != nil {
				ctx.AbortWithStatus(ir.Status)
				return ir.Err
			}
			Invocation2HTTPRequest(inv, ctx)

			// check body size
			if opts.bodyLimit > 0 {
				ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, opts.bodyLimit)
			}

			ctx.Next() //process user's handlers

			ir.Status = ctx.Writer.Status()
			if ir.Status >= http.StatusBadRequest {
				errMsg := ctx.Errors.ByType(gin.ErrorTypePrivate).String()
				if errMsg != "" {
					ir.Err = fmt.Errorf(errMsg)
				} else {
					ir.Err = fmt.Errorf("get err from http handle, get status: %d", ir.Status)
				}
			}
			ir.KVs = ctx.Keys
			ir.RequestID = requestid.FromContext(inv.Ctx)
			return ir.Err
		})
	}
}

func bindHc(gs *gin.Engine, opts *initOption) error {
	if !config.Get().Healthy.HcDisabled {
		hcPath := config.Get().Healthy.HcPath
		if !strings.HasPrefix(hcPath, "/") {
			hcPath = "/" + hcPath
		}
		qlog.Info("enable hc API on " + hcPath)

		gs.GET(hcPath, func(ctx *gin.Context) {
			type response struct {
				Status    string            `json:"status"`
				Substatus map[string]string `json:"substatus"`
			}

			substatus := make(map[string]string)
			if err := redis.CheckValid(); err == nil {
				substatus["redis"] = "success"
			} else {
				substatus["redis"] = "fail"
			}
			if err := gorm.CheckValid(); err == nil {
				substatus["mysql"] = "success"
			} else {
				substatus["mysql"] = "fail"
			}

			var extraResult map[string]string
			if opts.extraHc != nil {
				extraResult = opts.extraHc()
			}
			if extraResult != nil {
				for k, v := range extraResult {
					substatus[k] = v
				}
			}

			r := &response{
				Status:    "success",
				Substatus: substatus,
			}
			for _, v := range substatus {
				if v == "fail" {
					r.Status = "fail"
					break
				}
			}

			ctx.JSON(http.StatusOK, r)
		})
	}

	return nil
}

// HTTPRequest2Invocation convert http request to uniform invocation data format
func HTTPRequest2Invocation(ctx *gin.Context, serverName, schema, operation string) (*invocation.Invocation, error) {
	inv := &invocation.Invocation{
		MicroServiceName:   serverName,
		SourceMicroService: common.GetXQMSContext(common.HeaderSourceName, ctx.Request),
		Args:               ctx.Request,
		Protocol:           protocol.ProtocHTTP,
		Env:                qenv.Get(),
		SchemaID:           schema,
		OperationID:        operation,
		URLPathFormat:      ctx.Request.URL.Path,
		BriefURI:           ctx.FullPath(),
		Metadata: map[string]interface{}{
			common.RestMethod: ctx.Request.Method,
		},
	}
	if inv.BriefURI == "" { // 无效路由会导致 briefURI为空的
		inv.BriefURI = "/404"
		qlog.Debugf("not found route: %s", inv.URLPathFormat)
	}

	//set headers to Ctx, then user do not  need to consider about protocol in handlers
	header := common.Header{}
	inv.Ctx = common.NewContext(context.Background(), header)
	for k, vv := range ctx.Request.Header {
		header.Set(k, vv...)
	}
	return inv, nil
}

func (r *ginServer) Register(schema interface{}, opts ...server.Option) (string, error) {
	qlog.Info("register rest server(gin)")
	return "", nil
}

// Invocation2HTTPRequest convert invocation back to http request, set down all meta data
func Invocation2HTTPRequest(inv *invocation.Invocation, ctx *gin.Context) {
	for k, v := range inv.Metadata {
		ctx.Set(k, v.(string))
	}
	// set header
	contextHeader := common.FromContext(inv.Ctx)
	ctx.Request.Header = httputil.MergeHttpHeader(ctx.Request.Header, contextHeader)

	ctx.Request = ctx.Request.WithContext(inv.Ctx)

}

func (r *ginServer) Start() error {
	var err error
	config := r.opts
	r.mux.Lock()
	r.opts.address = config.address
	r.mux.Unlock()

	if r.opts.tLSConfig != nil {
		r.server = &http.Server{
			Addr:         config.address,
			Handler:      r.gs,
			TLSConfig:    r.opts.tLSConfig,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: time.Minute,
			IdleTimeout:  time.Minute,
		}
	} else {
		r.server = &http.Server{
			Addr:         config.address,
			Handler:      r.gs,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: time.Minute,
			IdleTimeout:  time.Minute,
		}
	}

	listen := r.opts.listen
	if listen == nil {
		//l, lIP, lPort, err := iputil.StartListener(config.address, config.tLSConfig)
		l, _, _, err := iputil.StartListener(config.address, config.tLSConfig)
		if err != nil {
			return fmt.Errorf("failed to start listener: %s", err.Error())
		}
		listen = l
		//registry.InstanceEndpoints[config.serverName] = net.JoinHostPort(lIP, lPort)
	}

	if laddr := listen.Addr().String(); !iputil.MatchServerPort(laddr, config.address) {
		qlog.Panicf("服务端口不匹配，想要[%s]，实际[%s]", config.address, laddr)
	}

	go func() {
		err = r.server.Serve(listen)
		if err != nil {
			qlog.Warn("http server err: " + err.Error())
			server.ErrRuntime <- err
		}
	}()

	qlog.Infof("%s server listening on: %s", r.opts.serverName, listen.Addr())
	return nil
}

func (r *ginServer) Stop() error {
	if r.server == nil {
		qlog.Info("http server never started")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r.server.SetKeepAlivesEnabled(false)

	//only golang 1.8 support graceful shutdown.
	if err := r.server.Shutdown(ctx); err != nil {
		qlog.Warn("http shutdown error: " + err.Error())
		return err // failure/timeout shutting down the server gracefully
	}
	return nil
}

func (r *ginServer) String() string {
	return Name
}

func (r *ginServer) Engine() interface{} {
	return r.gs
}
