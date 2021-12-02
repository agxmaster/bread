package validator

type Validator interface {
	Validate(v interface{}) error
	TransferCondition(table, field string, v interface{}, params map[string]interface{}) []Condition
}

type Condition interface {
	TableName() string
	FieldName() string
	Operator() string
	GetFullField() string
	GetSql() string
	ConditionValue() []interface{}
}
