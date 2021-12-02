package condition

import "reflect"

type QueryParam struct {
	*DBParam
	operator       string
	conditionValue []interface{}
	sql            string
}

func (q *QueryParam) TableName() string {
	return q.Table
}

func (q *QueryParam) FieldName() string {
	return q.Field
}
func (q *QueryParam) Operator() string {
	return q.operator
}

func (q *QueryParam) GetSql() string {
	return q.sql
}

func (q *QueryParam) ConditionValue() []interface{} {
	return q.conditionValue
}

func NewQueryParam(table, field, operator string, value interface{}) *QueryParam {
	return &QueryParam{
		DBParam:        newDatabaseParam(table, field),
		operator:       operator,
		conditionValue: []interface{}{value},
	}
}

func NewQueryParamWithSql(table, field, sql string, value []interface{}) *QueryParam {
	return &QueryParam{
		DBParam:        newDatabaseParam(table, field),
		sql:            sql,
		conditionValue: value,
	}
}

func NewDefaultQueryParam(table, field string, value interface{}) *QueryParam {
	rt := reflect.ValueOf(value)
	var operator string
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		operator = "in"
	default:
		operator = "="
	}
	return NewQueryParam(table, field, operator, value)
}
