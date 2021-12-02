package models
import "reflect"

var Student = studentModel{
	Table: NewTable("student", map[string]*Association{
		"class": {
			ForeignKey:  "id",
			LocalKey:    "class_id",
			TargetTable: "class",
			Join:        LeftJoin,
		},
	}),
	ID: TableField{
		Type:       reflect.Uint64,
		Name:       "id",
		Permission: Read,
	},
	Name: TableField{
		Type:       reflect.String,
		Name:       "name",
		Permission: Read,
	},
	Sex: TableField{
		Type:       reflect.Int,
		Name:       "sex",
		Permission: Read,
	},
	ClassId: TableField{
		Type:       reflect.Int,
		Name:       "class_id",
		Permission: Read,
	},
	CreateTime: TableField{
		Type:       reflect.String,
		Name:       "create_time",
		Permission: Read,
	},
}

type studentModel struct {
	ID                   TableField
	Name            	 TableField
	Sex           		 TableField
	ClassId              TableField
	CreateTime   		 TableField
	Table
}

