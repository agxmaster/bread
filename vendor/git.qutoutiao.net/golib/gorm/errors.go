package gorm

import (
	"errors"

	gormio "gorm.io/gorm"
)

// errors
var (
	ErrNotFoundClient = errors.New("no client available")
	ErrNotFoundConfig = errors.New("no config found")
	ErrInvalidConfig  = errors.New("named config is not a valid *Config type")

	// ErrRecordNotFound record not found error
	ErrRecordNotFound = gormio.ErrRecordNotFound
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gormio.ErrInvalidTransaction
	// ErrNotImplemented not implemented
	ErrNotImplemented = gormio.ErrNotImplemented
	// ErrMissingWhereClause missing where clause
	ErrMissingWhereClause = gormio.ErrMissingWhereClause
	// ErrUnsupportedRelation unsupported relations
	ErrUnsupportedRelation = gormio.ErrUnsupportedRelation
	// ErrPrimaryKeyRequired primary keys required
	ErrPrimaryKeyRequired = gormio.ErrPrimaryKeyRequired
	// ErrModelValueRequired model value required
	ErrModelValueRequired = gormio.ErrModelValueRequired
	// ErrInvalidData unsupported data
	ErrInvalidData = gormio.ErrInvalidData
	// ErrUnsupportedDriver unsupported driver
	ErrUnsupportedDriver = gormio.ErrUnsupportedDriver
	// ErrRegistered registered
	ErrRegistered = gormio.ErrRegistered
	// ErrInvalidField invalid field
	ErrInvalidField = gormio.ErrInvalidField
	// ErrEmptySlice empty slice found
	ErrEmptySlice = gormio.ErrEmptySlice
	// ErrDryRunModeUnsupported dry run mode unsupported
	ErrDryRunModeUnsupported = gormio.ErrDryRunModeUnsupported
)
