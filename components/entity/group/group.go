package group

import (
	"reflect"
	"sync/atomic"

	"github.com/go-bread/components/entity/field"
	"github.com/go-bread/consts"

	"github.com/go-bread/components/entity/models"
)

type FieldsMap map[consts.EntityGroupName]EntityGroup

// entity group
type EntityGroup struct {
	JoinDriveTable  models.Table // 关联驱动表
	Entities        map[string]interface{}
	loadedAllFields int32
	dividedFields   map[string][]string
}

func (e *EntityGroup) Init() {
	if atomic.CompareAndSwapInt32(&e.loadedAllFields, 0, 1) {
		divide := make(map[string][]string)
		for _, v := range e.Entities {
			f, ok := v.(field.Field)
			if !ok {
				continue
			}
			if _, ok := divide[f.Table.TableName()]; !ok {
				var fields []string
				table := f.Table
				t := reflect.ValueOf(table)
				pv := reflect.Indirect(t)
				tp := pv.Type()
				for i := 0; i < pv.NumField(); i++ {
					f := pv.Field(i)
					if tp.Field(i).Name == "Table" {
						continue
					}
					if f.Interface().(models.TableField).Name == "" {
						continue
					}
					fields = append(fields, f.Interface().(models.TableField).Name)
				}
				divide[f.Table.TableName()] = fields
			}
		}
		e.dividedFields = divide
	}
}

func (e *EntityGroup) Initialized() bool {
	return atomic.LoadInt32(&e.loadedAllFields) != 0
}

func (e *EntityGroup) GetDividedEntities() map[string][]string {
	if atomic.LoadInt32(&e.loadedAllFields) == 0 {
		panic("uninitialized entity group")
	}
	return e.dividedFields
}
