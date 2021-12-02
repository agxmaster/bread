package rest

import (
	"crypto/tls"
)

type CallOption struct {
	isTls   bool
	failure map[string]bool
}

func NewCallOption(tLSConfig *tls.Config, failure map[string]bool) *CallOption {
	opts := &CallOption{}
	if tLSConfig != nil {
		opts.isTls = true
	}

	if failure != nil {
		opts.failure = failure
	}
	return opts
}
