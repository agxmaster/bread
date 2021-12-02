package gorm

import (
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// defaults
const (
	DefaultDriver   = "mysql"
	MaxDialTimeout  = 1000 // millisecond
	MaxReadTimeout  = 3000 // millisecond
	MaxWriteTimeout = 3000 // millisecond
	MaxOpenConn     = 200
	MaxIdleConn     = 80
	MaxLifetime     = 600 // in second
)

type (
	Field      = schema.Field
	Expression = clause.Expression
	Writer     = clause.Writer
)
