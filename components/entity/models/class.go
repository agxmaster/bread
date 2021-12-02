package models

import (
	"reflect"
)

var (
	Class = classModel{
		Id: TableField{
			Type:       reflect.Uint64,
			Name:       "id",
			Permission: Read,
		},
		ClassName: TableField{
			Type:       reflect.String,
			Name:       "class_name",
			Permission: Read,
		},
		CreateTime: TableField{
			Type:       reflect.Uint8,
			Name:       "create_time",
			Permission: Read,
		},
		Table: NewTable("class", nil),
	}
)

type classModel struct {
	Id            		 TableField
	ClassName            TableField
	CreateTime	         TableField
	Table
}
