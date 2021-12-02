package models

import "reflect"

type table struct {
	name         string                  // 表名
	primaryKey   string                  // 主键
	associations map[string]*Association // 关联关系
}

func NewTable(name string, associations map[string]*Association) Table {
	return table{
		name:         name,
		associations: associations,
	}
}

func (t table) TableName() string {
	return t.name
}

func (t table) GetAssociation(name string) *Association {
	if v, ok := t.associations[name]; ok {
		return v
	}
	return nil
}

func (t table) PrimaryKey() string {
	return t.primaryKey
}

type Table interface {
	TableName() string
	GetAssociation(name string) *Association
	PrimaryKey() string
}

type JoinMethod string

var (
	LeftJoin  JoinMethod = "LEFT JOIN"
	RighJoin  JoinMethod = "RIGHT JOIN"
	InnerJoin JoinMethod = "INNER JOIN"
)

type Association struct {
	ForeignKey  string
	LocalKey    string
	TargetTable string
	Join        JoinMethod
}

type Permission string

const (
	Read      Permission = "r"
	ReadWrite Permission = "rw"
)

type TableField struct {
	Type       reflect.Kind
	Name       string
	Permission Permission
}
