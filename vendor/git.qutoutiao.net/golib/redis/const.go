package redis

import redisv8 "github.com/go-redis/redis/v8"

const (
	Nil = redisv8.Nil
)

const (
	MaxDialTimeout  = 1000 // millisecond
	MaxReadTimeout  = 100  // millisecond
	MaxWriteTimeout = 100  // millisecond
	MaxPoolSize     = 200
	MaxPoolTimeout  = MaxReadTimeout + 1000 // millisecond
	MinIdleConns    = 30
	MaxRetries      = 0

	redisSuccess = "200"
	redisFailed  = "400"
)
