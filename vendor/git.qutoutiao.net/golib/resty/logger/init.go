package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var (
	zlog = zerolog.New(os.Stderr).With().Timestamp().Logger()
)
