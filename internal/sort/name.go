package sort

import (
	"sort"
	"strings"

	"github.com/ipanardian/lu-hut/internal/model"
)

type Name struct{}

func (s *Name) Sort(files []model.FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		result := strings.Compare(strings.ToLower(files[i].Name), strings.ToLower(files[j].Name))
		if reverse {
			return result > 0
		}
		return result < 0
	})
}
