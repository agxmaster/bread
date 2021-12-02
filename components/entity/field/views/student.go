package views

import (
	"github.com/go-bread/components/entity/field"
	"github.com/go-bread/components/entity/group"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-bread/components/entity/models"
	"github.com/go-bread/iface/entity_query"
)

var (
	Student = group.EntityGroup{
		JoinDriveTable: models.Student,
		Entities: map[string]interface{}{
			"id": field.Field{
				Table:      models.Student,
				TableField: models.Student.ID,
				CanQuery:   true,
			},
			"name": field.Field{
				Table:      models.Student,
				TableField: models.Student.Name,
			},
			"sex": field.Field{
				Table:      models.Student,
				TableField: models.Student.Sex,
				CanQuery:   true,
			},
			"class_id": field.Field{
				Table:      models.Student,
				TableField: models.Student.ClassId,
				CanQuery:   true,
			},
			"class_name": field.Field{
				Table:      models.Class,
				TableField: models.Class.ClassName,
				CanQuery:   true,
			},
			"create_time": field.Field{
				Table:      models.Student,
				TableField: models.Student.CreateTime,
				CanQuery:   true,
				Callback: func(ctx *gin.Context, i interface{}, values map[string]interface{}, ls *entity_query.LocalStorage) interface{} {
					if t, ok := i.(time.Time); ok {
						return t.Format("2006-01-02 15:04:05")
					}
					if t, ok := i.(*time.Time); ok {
						return t.Format("2006-01-02 15:04:05")
					}
					return i
				},
			},
		},
	}
)

