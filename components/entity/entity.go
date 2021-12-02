package entity

import (
	"fmt"
	"github.com/go-bread/components/entity/field/views"
	"github.com/pkg/errors"

	"github.com/go-bread/components/database"
	outputs "github.com/go-bread/components/database/output"
	"github.com/go-bread/components/entity/field"
	"github.com/go-bread/components/entity/group"
	"github.com/go-bread/consts"
	validatorIface "github.com/go-bread/iface/validator"
	"github.com/go-bread/validators/query"
	"github.com/gin-gonic/gin"
)

type outputFields struct {
	List []map[string]interface{} `json:"list"`
	Page query.Pagination         `json:"page_info"`
}

var (

	FieldsMap = group.FieldsMap{
		consts.EntityStudent: views.Student,
		consts.EntityClass: views.Class,
	}
)

const (
	SceneUpdate = "update"
	SceneQuery  = "query"
	SceneCreate = "create"
)

func QueryAndFormatOne(ctx *gin.Context, fieldsMap group.FieldsMap, gn consts.EntityGroupName, params query.QParams) (map[string]interface{}, error) {
	data, err := parseAndQueryAll(ctx, fieldsMap, gn, &params)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}
	dealValue(data[0])

	return data[0], nil
}

func QueryAndFormatAll(ctx *gin.Context, fieldsMap group.FieldsMap, gn consts.EntityGroupName, params query.QParams) (interface{}, error) {
	data, err := parseAndQueryAll(ctx, fieldsMap, gn, &params)
	if err != nil {
		return nil, err
	}

	var r outputFields
	if data == nil {
		r.List = []map[string]interface{}{}
	} else {
		for _, v := range data {
			dealValue(v)
		}
		r.List = data
	}
	r.Page = params.Pagination

	return r, nil
}

func parseAndQueryAll(ctx *gin.Context, fieldsMap group.FieldsMap, gn consts.EntityGroupName, params *query.QParams) ([]map[string]interface{}, error) {
	fm, ok := fieldsMap[gn]

	if !ok {
		return nil, 	errors.New("invalid group")
	}
	if !fm.Initialized() {
		fm.Init()
		fieldsMap[gn] = fm
	}

	// 参数校验
	var queryParams []validatorIface.Condition
	queryParams, err := ValidateAndBuildParams(queryParams, SceneQuery, fm.Entities, params.QFields)
	if err != nil {
		return nil, err
	}

	// 输出字段校验及构建
	outputFields, err := ValidateAndBuildOutputs(fm, params.ReturnFields)
	if err != nil {
		return nil, err
	}

	// order by 处理
	orders, err := validateAndBuildOrders(fm, params.Orders)
	if err != nil {
		return nil, err
	}

	// 执行查询
	q, err := database.QueryAndFormat(ctx, fieldsMap[gn], queryParams, outputFields, &params.Pagination, orders)
	if err != nil {
		return nil, err
	}

	return q, nil
}

// validate input params and build db params
func ValidateAndBuildParams(queryParams []validatorIface.Condition, scene string, groupFields map[string]interface{}, params map[string]interface{}) ([]validatorIface.Condition, error) {
	for p, v := range params {
		f, ok := groupFields[p]
		// 检测字段是否存在
		if !ok {
			return nil, errors.New("invalid params, "+ p+"字段不存在")
		}

		if ff, ok := f.(field.Field); ok {
			ff.InputField = p
			err := validateFieldValue(v, ff)
			if err != nil {
				return nil, err
			}
			if !ff.CanQuery {
				continue
			}
			queryParams = append(queryParams, ff.TransferCondition(v, params)...)
		} else if ff, ok := f.(map[string]interface{}); ok {
			vv, ok := v.(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid params, " + p + "实体不存在")
			}

			return ValidateAndBuildParams(queryParams, scene, ff, vv)
		} else {
			panic("wrong fields map: key " + p)
		}
	}

	return queryParams, nil
}

func ValidateAndBuildOutputs(g group.EntityGroup, fields []string) ([]*outputs.OutputField, error) {
	var ops []*outputs.OutputField
	// 用于字段去重
	fp := make(map[string]struct{})
	for _, k := range fields {
		if _, ok := fp[k]; ok {
			continue
		}
		fp[k] = struct{}{}

		f, ok := g.Entities[k]
		if !ok {
			return nil, errors.New("invalid field, " + fmt.Sprintf("字段%s不存在", k))
		}

		if ff, ok := f.(field.Field); ok {
			ops = append(ops, &outputs.OutputField{
				TableField: ff.TableField.Name,
				Table:      ff.Table.TableName(),
				OutPut:     k,
				F:          ff.Callback,
			})
		}

		if ff, ok := f.(map[string]field.Field); ok {
			_, ok := ff[k]
			if !ok {
				return nil, errors.New("invalid field, " + fmt.Sprintf("字段%s不存在", k))
			}

			if ff, ok := f.(field.Field); ok {
				ops = append(ops, &outputs.OutputField{
					TableField: ff.TableField.Name,
					Table:      ff.Table.TableName(),
					OutPut:     k,
					F:          ff.Callback,
				})
			}
		}
	}
	return ops, nil
}

// 构建order by
func validateAndBuildOrders(g group.EntityGroup, orders [][2]string) ([][2]string, error) {
	if len(orders) == 0 {
		return [][2]string{}, nil
	}
	var formatedOrders [][2]string
	// 用于字段去重
	fp := make(map[string]struct{})
	for _, k := range orders {
		if _, ok := fp[k[0]]; ok {
			continue
		}
		fp[k[0]] = struct{}{}

		f, ok := g.Entities[k[0]]
		if !ok {
			return nil, errors.New("invalid field, " + fmt.Sprintf("排序字段%s不存在", k[0]))
		}

		if ff, ok := f.(field.Field); ok {
			if !ff.CanOrder {
				return nil, errors.New("invalid field, " + fmt.Sprintf("不能使用字段%s进行排序", k[0]))
			}
			formatedOrders = append(formatedOrders, [2]string{
				fmt.Sprintf("%s.%s", ff.Table.TableName(), ff.TableField.Name),
				k[1],
			})
		}
	}
	return formatedOrders, nil
}

// use field.Field.Validators to validate the value
func validateFieldValue(v interface{}, f field.Field) error {
	if !f.CanQuery {
		return errors.New("invalid field, " + fmt.Sprintf("%s字段不能作为查询条件", f.InputField))
	}

	if f.Validator == nil {
		return nil
	}

	err := f.Validator.Validate(v)
	if err != nil {
		return err
	}

	return nil
}

func dealValue(m map[string]interface{}) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
}
