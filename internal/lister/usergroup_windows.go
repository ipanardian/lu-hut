//go:build windows

package lister

import (
	"os"

	"github.com/fatih/color"
)

// extractUserGroup returns "unknown" on Windows, where os.FileInfo.Sys() does
// not expose POSIX owner/group metadata.
func extractUserGroup(fileInfo os.FileInfo) (string, string) {
	return color.New(color.FgWhite).Sprint("unknown"), color.New(color.FgWhite).Sprint("unknown")
}
