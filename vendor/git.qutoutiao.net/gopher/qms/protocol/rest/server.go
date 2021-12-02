package rest

import (
	"fmt"

	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"git.qutoutiao.net/gopher/qms/internal/engine"
	"github.com/gin-gonic/gin"
)

//Gin 返回*gin.Engine对象，使用方可以用于注册路由，具体用法完全等同于Gin。
func Gin(opts ...ServerOption) (*gin.Engine, error) {
	s, err := engine.GetServer("rest", opts...)
	if err != nil {
		return nil, err
	}
	if ginServer, ok := s.(server.GinServer); ok {
		return ginServer.Engine().(*gin.Engine), nil
	}
	return nil, fmt.Errorf("server(%s) is not implemented with gin", s.String())
}
