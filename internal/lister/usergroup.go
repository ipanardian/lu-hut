package lister

import (
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/fatih/color"
)

func extractUserGroup(fileInfo os.FileInfo) (string, string) {
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		u, errU := user.LookupId(strconv.Itoa(int(stat.Uid)))
		g, errG := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

		username := "unknown"
		groupname := "unknown"

		if errU == nil {
			username = u.Username
		}
		if errG == nil {
			groupname = g.Name
		}

		return color.New(color.FgWhite).Sprint(username), color.New(color.FgWhite).Sprint(groupname)
	}
	return color.New(color.FgWhite).Sprint("unknown"), color.New(color.FgWhite).Sprint("unknown")
}
