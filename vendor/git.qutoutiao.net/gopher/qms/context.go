package qms

import (
	"context"

	"git.qutoutiao.net/gopher/qms/internal/core/requestid"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qms/protocol/rest/render"
	"github.com/gin-gonic/gin"
	grender "github.com/gin-gonic/gin/render"
)

type GinContext = gin.Context

// Context是qms框架的统一context
type Context struct {
	*GinContext
	ctx    context.Context
	logger qlog.Logger
}

// HandlerFunc defines the handler used by qms middleware as return value.
type HandlerFunc func(*Context)

// NewContextFromGin 实现由gin.Context转换为qms.Context
func NewContextFromGin(c *gin.Context) *Context {
	return &Context{
		GinContext: c,
		ctx:        c.Request.Context(),
		logger:     qlog.GetLogger(),
	}
}

// NewContext 实现由go标准Context转换为qms.Context
func NewContext(c context.Context) *Context {
	return &Context{
		GinContext: &gin.Context{},
		ctx:        c,
		logger:     qlog.GetLogger(),
	}
}

// Context 返回go标准Context
func (c Context) Context() context.Context {
	return c.ctx
}

// RequestID 返回request_id（与traceid是同一个值）
func (c Context) RequestID() string {
	return requestid.FromContext(c.ctx)
}

// Logger 返回携带request_id的Logger
func (c Context) Logger() qlog.Logger {
	return c.logger.WithField("request_id", c.RequestID())
}

// WrapGinHandler 由qms.HandlerFunc转为gin.HandlerFunc
func WrapGinHandler(handlerFunc HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		qmsCtx := NewContextFromGin(ctx)
		handlerFunc(qmsCtx)
	}
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r grender.Render) {
	defer func() {
		if err := recover(); err != nil {
			qlog.Errorf("code: %d, err: %v", code, err)
		}
	}()

	c.GinContext.Render(code, r)
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, grender.JSON{Data: obj})
}

// PureJSON serializes the given struct as JSON into the response body.
// PureJSON, unlike JSON, does not replace special html characters with their unicode entities.
func (c *Context) PureJSON(code int, obj interface{}) {
	c.Render(code, grender.PureJSON{Data: obj})
}

// JSONPB 自定义marshal解析由proto定义的json数据
func (c *Context) JSONPB(code int, obj interface{}) {
	c.Render(code, render.Jsonpb{Data: obj})
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) XML(code int, obj interface{}) {
	c.Render(code, grender.XML{Data: obj})
}

// YAML serializes the given struct as YAML into the response body.
func (c *Context) YAML(code int, obj interface{}) {
	c.Render(code, grender.YAML{Data: obj})
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, grender.String{Format: format, Data: values})
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, grender.Data{
		ContentType: contentType,
		Data:        data,
	})
}
