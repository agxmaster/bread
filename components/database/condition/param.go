package condition

type DBParam struct {
	Table string
	Field string
}

func (p *DBParam) GetFullField() string {
	return p.Table + "." + p.Field
}

func newDatabaseParam(table, field string) *DBParam {
	return &DBParam{
		Table: table,
		Field: field,
	}
}
