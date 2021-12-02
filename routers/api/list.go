package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-bread/components/entity"
	"github.com/go-bread/consts"
	"github.com/go-bread/validators/query"
	"net/http"
)

func GetList(c *gin.Context) {

	qp, _ := query.Parse(c)
	form := c.Param("form")
	respData, _ := entity.QueryAndFormatAll(c, entity.FieldsMap, consts.EntityGroupName(form), qp)

	c.JSON(http.StatusOK,respData)
}
