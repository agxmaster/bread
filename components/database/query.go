package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-bread/components/entity/group"
	"github.com/gin-gonic/gin"

	outputs "github.com/go-bread/components/database/output"
	models2 "github.com/go-bread/components/entity/models"
	"github.com/go-bread/iface/entity_query"
	validatorIface "github.com/go-bread/iface/validator"
	"github.com/go-bread/models"
	"github.com/go-bread/validators/query"
)

const (
	// logical operator
	Equal           = "="
	NotEqual        = "!="
	NotEqual2       = "<>"
	LargerThan      = ">"
	LargerEqualThan = ">="
	LessThan        = "<"
	LessEqualThan   = "<="
	In              = "in"
	Like            = "like"

	And = "and"
	Or  = "or"
)

func QueryAndFormat(ctx *gin.Context, group group.EntityGroup, params []validatorIface.Condition, outputs []*outputs.OutputField, pagination *query.Pagination, order [][2]string) ([]map[string]interface{}, error) {
	var r []map[string]interface{}
	// 回调函数处理
	callbacks := callbackBuild(outputs)
	// 查询参数处理
	cond, vals, err := whereBuild(params)
	if err != nil {
		return nil, err
	}

	tables := uniqueTables(outputs)

	// select字段处理
	selectFields := selectFieldsBuild(group, tables)

	// 多表关联查询
	associations := make(map[string]*models2.Association)
	var majorTable string
	if len(tables) > 1 {
		if len(tables) > 2 {
			panic(fmt.Sprintf("暂不支持超过两张以上表关联查询 %+v", tables))
		}
		if group.JoinDriveTable == nil {
			panic("多表关联查询必须声明驱动表")
		}
		majorTable = group.JoinDriveTable.TableName()
		for t := range tables {
			if t == group.JoinDriveTable.TableName() {
				continue
			}
			ass := group.JoinDriveTable.GetAssociation(t)
			if ass == nil {
				panic(fmt.Sprintf("关联关系未定义: majorTable: %s, target: %s", group.JoinDriveTable.TableName(), t))
			}
			if _, ok := associations[t]; !ok {
				associations[t] = ass
			}
		}
	} else {
		for v := range tables {
			majorTable = v
			break
		}
	}

	model := models.GetDb().Table(majorTable)
	model.LogMode(true)
	if len(associations) > 0 {
		for _, ass := range associations {
			model = model.Joins(fmt.Sprintf("%s %s on %s.%s = %s.%s", ass.Join, ass.TargetTable, ass.TargetTable, ass.ForeignKey, group.JoinDriveTable.TableName(), ass.LocalKey))
		}
	}
	model = model.Where(cond, vals...)
	if pagination.Page != 0 {
		model.Count(&pagination.TotalCount)
		model = model.Offset((pagination.Page - 1) * pagination.PageSize).Limit(pagination.PageSize)
	}

	if len(order) > 0 {
		for _, v := range order {
			stringOrder := fmt.Sprintf("%s %s", v[0], v[1])
			model = model.Order(stringOrder)
		}
	}
	if len(selectFields) > 0 {
		model = model.Select(selectFields)
	}
	rows, err := model.Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Print("close err:%s", closeErr.Error())
		}
	}()

	columns, err := rows.Columns()
	length := len(columns)
	if err != nil {
		return nil, err
	}
	var finalRows []map[string]interface{}
	primaryKeys := make([]uint64, 0)
	for rows.Next() {
		current := makeResultReceiver(length)
		if err := rows.Scan(current...); err != nil {
			panic(err)
		}
		row := make(map[string]interface{})
		for i := 0; i < length; i++ {
			key := columns[i]
			val := *(current[i]).(*interface{})
			if val == nil {
				row[key] = nil
				continue
			}
			row[key] = val
			keys := strings.Split(key, ".")
			if id, ok := val.(int64); ok && keys[len(keys)-1] == "id" {
				primaryKeys = append(primaryKeys, uint64(id))
			}
		}
		finalRows = append(finalRows, row)
	}

	ls := entity_query.NewLS()
	// 存储当前页的所有主键, 部分场景下做数据预加载
	if len(primaryKeys) > 0 {
		ls.Set("primary_keys", primaryKeys)
	}
	for _, row := range finalRows {
		value := make(map[string]interface{})
		for _, o := range outputs {
			v, ok := row[fieldIndex(o)]
			if !ok {
				continue
			}
			outputKey, outputVal := formatValue(ctx, v, o, callbacks, ls, row)
			value[outputKey] = outputVal
		}

		r = append(r, value)
	}

	return r, nil
}

func formatValue(ctx *gin.Context, v interface{}, o *outputs.OutputField, c outputs.Callbacks, storage *entity_query.LocalStorage, row map[string]interface{}) (string, interface{}) {
	if f, ok := c[fieldCallbackIndex(o)]; ok {
		return o.OutPut, f(ctx, v, row, storage)
	}

	switch v.(type) {
	case []byte:
		return o.OutPut, string(v.([]byte))
	case time.Time:
		return o.OutPut, v.(time.Time).Format("2006-01-02 15:04:05")
	case *time.Time:
		return o.OutPut, v.(*time.Time).Format("2006-01-02 15:04:05")
	default:
		return o.OutPut, v
	}
}

func makeResultReceiver(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{}
		result = append(result, &current)
	}
	return result
}

func uniqueTables(op []*outputs.OutputField) map[string]bool {
	tables := make(map[string]bool)
	for _, v := range op {
		if _, ok := tables[v.Table]; !ok {
			tables[v.Table] = true
		}
	}

	return tables
}

// sql build where
func whereBuild(params []validatorIface.Condition) (whereSQL string, vals []interface{}, err error) {
	for _, v := range params {
		if whereSQL != "" {
			whereSQL += " AND "
		}

		// 如果指定了sql模板, 优先使用sql
		if v.GetSql() != "" {
			whereSQL += fmt.Sprintf(" (%s) ", v.GetSql())
		} else {
			k := v.GetFullField()
			switch v.Operator() {
			case Equal:
				whereSQL += fmt.Sprint(k, " =? ")
			case LargerThan:
				whereSQL += fmt.Sprint(k, " >? ")
			case LargerEqualThan:
				whereSQL += fmt.Sprint(k, " >=? ")
			case LessThan:
				whereSQL += fmt.Sprint(k, " <? ")
			case LessEqualThan:
				whereSQL += fmt.Sprint(k, " <=? ")
			case NotEqual:
				whereSQL += fmt.Sprint(k, " !=? ")
			case NotEqual2:
				whereSQL += fmt.Sprint(k, " !=? ")
			case In:
				whereSQL += fmt.Sprint(k, " in (?)")
			case Like:
				whereSQL += fmt.Sprint(k, " like ? ")
			}
		}

		vals = append(vals, v.ConditionValue()...)
	}
	return
}

func selectFieldsBuild(group group.EntityGroup, tables map[string]bool) []string {
	var selectFields []string
	for k := range tables {
		es, ok := group.GetDividedEntities()[k]
		if !ok {
			panic("不支持的table " + k)
		}
		for _, v := range es {
			selectFields = append(selectFields, fmt.Sprintf("`%s`.`%s` as '%s.%s'", k, v, k, v))
		}
	}
	return selectFields
}

func callbackBuild(ofs []*outputs.OutputField) outputs.Callbacks {
	m := make(map[string]entity_query.CallbackFunc)
	for _, o := range ofs {
		if o.F != nil {
			m[fieldCallbackIndex(o)] = o.F
		}
	}
	return m
}

func fieldIndex(o *outputs.OutputField) string {
	return fmt.Sprintf("%s.%s", o.Table, o.TableField)
}

func fieldCallbackIndex(o *outputs.OutputField) string {
	return fmt.Sprintf("%s.%s->%s", o.Table, o.TableField, o.OutPut)
}
