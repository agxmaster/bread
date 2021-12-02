package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GinHandleFunc is a gin handler which can expose metrics in http server
func Handle(w http.ResponseWriter, req *http.Request) {
	promhttp.Handler().ServeHTTP(w, req)
}

func GinHandleFunc(ctx *gin.Context) {
	promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
