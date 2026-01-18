package sort

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/ipanardian/lu-hut/internal/model"
)

type Extension struct{}

func (s *Extension) Sort(files []model.FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		extI := strings.ToLower(filepath.Ext(files[i].Name))
		extJ := strings.ToLower(filepath.Ext(files[j].Name))
		if reverse {
			return extI > extJ
		}
		return extI < extJ
	})
}
