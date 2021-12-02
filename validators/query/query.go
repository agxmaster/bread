package query

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-bread/utils/validate"
	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 500
)

type QParams struct {
	QFields
	ReturnFields
	Pagination
	OrderBy
}

type Parameters struct {
	Fields   []string `validate:"required" json:"fields"`
	Page     uint32   `json:"_page"`
	PageSize uint32   `json:"_page_size"`
}

type QFields map[string]interface{}

type ReturnFields []string

type Pagination struct {
	Page       uint32 `json:"page"`
	PageSize   uint32 `json:"page_size"`
	TotalCount uint32 `json:"total_number"`
	Offset     uint32 `json:"-"`
}

type OrderBy struct {
	Orders [][2]string `json:"_order_by" validate:"omitempty,dive,len=2"`
}



func Parse(ctx *gin.Context) (QParams, error) {
	// build query params
	var qp QParams
	q := ctx.Query("query")
	if q == "" {
		return qp, errors.New("query field is required")
	}

	var parameters Parameters
	err := json.Unmarshal([]byte(q), &parameters)
	if err != nil {
		return qp, err
	}

	err = validate.StructParam(parameters)
	if err != nil {
		return qp, err
	}

	var p Pagination
	// init page info
	if parameters.Page != 0 {
		p.Page = parameters.Page
		p.PageSize = parameters.PageSize
		p.Init()
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(q), &m)
	if err != nil {
		return qp, errors.New("query field is invalid")
	}

	var orders OrderBy
	if len(orders.Orders) > 0 {
		err = json.Unmarshal([]byte(q), &orders)
		if err != nil {
			return qp, errors.New("order by is invalid")
		}
		err = validate.StructParam(orders)
		if err != nil {
			return qp, err
		}
		for i, v := range orders.Orders {
			if strings.ToLower(v[1]) != "asc" && strings.ToLower(v[1]) != "desc" {
				return qp, errors.New("order by is invalid")
			}
			orders.Orders[i] = [2]string{v[0], strings.ToLower(v[1])}
		}
	}

	delete(m, "fields")
	delete(m, "page")
	delete(m, "page_size")
	delete(m, "order_by")
	return QParams{
		QFields:      m,
		ReturnFields: parameters.Fields,
		Pagination:   p,
		OrderBy:      orders,
	}, nil
}

func (p *Pagination) Init() {
	if p.Page == 0 {
		p.Page = DefaultPage
	}
	if p.PageSize == 0 || p.PageSize > MaxPageSize {
		p.PageSize = DefaultPageSize
	}
	p.Offset = (p.Page - 1) * p.PageSize
}
