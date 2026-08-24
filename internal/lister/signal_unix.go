//go:build !windows

package lister

import (
	"os"
	"syscall"
)

func terminationSignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM}
}
