package icons

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeAlways Mode = "always"
	ModeNever  Mode = "never"
)

func ShouldShow(mode Mode) bool {
	switch mode {
	case ModeAlways:
		return true
	case ModeNever:
		return false
	default:
		return isInteractiveTerminal()
	}
}

func isInteractiveTerminal() bool {
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

const IconWidth = 2

func ForFile(name string, isDir bool, isEmpty bool) rune {
	if isDir {
		if icon, ok := directoryIcons[name]; ok {
			return icon
		}
		if isEmpty {
			return iconFolderOpen
		}
		return iconFolder
	}

	if icon, ok := filenameIcons[name]; ok {
		return icon
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if ext == "" {
		return iconFileUnknown
	}
	if icon, ok := extensionIcons[ext]; ok {
		return icon
	}
	return iconFile
}

func Prefix(name string, isDir bool, isEmpty bool, mode Mode) string {
	if !ShouldShow(mode) {
		return ""
	}
	return string(ForFile(name, isDir, isEmpty)) + " "
}
