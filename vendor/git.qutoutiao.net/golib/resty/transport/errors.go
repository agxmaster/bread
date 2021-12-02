package transport

import "errors"

var (
	ErrResolverError    = errors.New("cannot resolve service")
	ErrServiceNotFound  = errors.New("service is not found")
	ErrInvalidTransport = errors.New("transport is invalid")
)
