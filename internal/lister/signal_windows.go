//go:build windows

package lister

import "os"

func terminationSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
