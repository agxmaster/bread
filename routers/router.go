package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-bread/routers/api"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.GET("list/:form",api.GetList)
	r.GET("create/:form",api.Create)

	return r
}
