package field

import (
	"github.com/go-bread/components/database/condition"
	"github.com/go-bread/components/entity/models"
	"github.com/go-bread/iface/entity_query"
	validatorIface "github.com/go-bread/iface/validator"
)

type Field struct {
	Table      models.Table
	TableField models.TableField
	Validator  validatorIface.Validator
	CanQuery   bool // 是否可以用来做查询
	CanOrder   bool // 是否可以用来排序
	InputField string
	Callback   entity_query.CallbackFunc
}

func (f *Field) TransferCondition(v interface{}, params map[string]interface{}) []validatorIface.Condition {
	if !f.CanQuery {
		return []validatorIface.Condition{}
	}

	if f.Validator == nil {
		return []validatorIface.Condition{condition.NewDefaultQueryParam(f.Table.TableName(), f.TableField.Name, v)}
	}

	return f.Validator.TransferCondition(f.Table.TableName(), f.TableField.Name, v, params)
}
