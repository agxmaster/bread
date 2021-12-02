package handler

import "errors"

var errViolateBuildIn = errors.New("can not replace build-in handler func")

//ErrDuplicatedHandler means you registered more than 1 handler with same name
var ErrDuplicatedHandler = errors.New("duplicated handler registration")
