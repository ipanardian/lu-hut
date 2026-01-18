package sort

import (
	"sort"

	"github.com/ipanardian/lu-hut/internal/model"
)

type Size struct{}

func (s *Size) Sort(files []model.FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		if reverse {
			return files[i].Size < files[j].Size
		}
		return files[i].Size > files[j].Size
	})
}
