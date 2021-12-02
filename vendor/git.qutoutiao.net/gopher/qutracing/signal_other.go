// +build !windows

package qutracing

import (
	"os"
	"syscall"
)

var DefaultSignal os.Signal = syscall.SIGUSR1
