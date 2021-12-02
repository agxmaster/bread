package views

import (
	"github.com/go-bread/components/entity/field"
	"github.com/go-bread/components/entity/group"
	"github.com/go-bread/components/entity/models"
)

var (
	Class = group.EntityGroup{
		Entities: map[string]interface{}{
			"id": field.Field{
				Table:      models.Class,
				TableField: models.Class.Id,
				CanQuery:   true,
			},
			"class_name": field.Field{
				Table:      models.Class,
				TableField: models.Class.ClassName,
			},
			"create_time": field.Field{
				Table:      models.Class,
				TableField: models.Class.CreateTime,
			},
		},
	}
)
