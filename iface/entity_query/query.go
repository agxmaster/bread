// nolint:golint
package entity_query

import "github.com/gin-gonic/gin"

type CallbackFunc func(*gin.Context, interface{}, map[string]interface{}, *LocalStorage) interface{}

type LocalStorage struct {
	l map[string]interface{}
}

func NewLS() *LocalStorage {
	return &LocalStorage{l: map[string]interface{}{}}
}

func (ls LocalStorage) IsSet(key string) bool {
	_, ok := ls.l[key]
	return ok
}

func (ls LocalStorage) Get(key string) interface{} {
	v, ok := ls.l[key]
	if !ok {
		return nil
	}
	return v
}

func (ls LocalStorage) Set(key string, value interface{}) {
	ls.l[key] = value
}
