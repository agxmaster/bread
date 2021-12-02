package rolling

import "time"

const (
	DefaultSize    = 10 // NOTE: DO NOT change this!!!
	DefaultTimeout = DefaultSize * time.Second
	DefaultLatency = time.Second
)
