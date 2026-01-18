// Package model defines data structures for file system entries.
package model

import (
	"io/fs"
	"time"
)

type FileEntry struct {
	Name      string
	Path      string
	Size      int64
	Mode      fs.FileMode
	ModTime   time.Time
	IsDir     bool
	IsHidden  bool
	GitStatus string
	Author    string
	Group     string
}
