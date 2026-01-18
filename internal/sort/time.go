package sort

import (
	"sort"

	"github.com/ipanardian/lu-hut/internal/model"
)

type Time struct{}

func (s *Time) Sort(files []model.FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if reverse {
			return files[i].ModTime.Before(files[j].ModTime)
		}
		return files[i].ModTime.After(files[j].ModTime)
	})
}
