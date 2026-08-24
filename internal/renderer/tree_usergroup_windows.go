//go:build windows

package renderer

import "os"

// extractUserGroup returns "unknown" on Windows, where os.FileInfo.Sys() does
// not expose POSIX owner/group metadata.
func (r *Tree) extractUserGroup(fileInfo os.FileInfo) (string, string) {
	return "unknown", "unknown"
}
