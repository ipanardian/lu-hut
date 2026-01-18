// Package sort provides strategies for sorting file entries.
package sort

import "github.com/ipanardian/lu-hut/internal/model"

type Strategy interface {
	Sort(files []model.FileEntry, reverse bool)
}
