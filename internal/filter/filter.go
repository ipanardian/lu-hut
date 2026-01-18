// Package filter provides functionality for filtering file entries based on patterns.
package filter

import (
	"path/filepath"

	"github.com/ipanardian/lu-hut/internal/model"
)

type Filter struct {
	includePatterns []string
	excludePatterns []string
}

func NewFilter(includePatterns, excludePatterns []string) *Filter {
	return &Filter{
		includePatterns: includePatterns,
		excludePatterns: excludePatterns,
	}
}

func (f *Filter) Apply(files []model.FileEntry, showHidden bool) []model.FileEntry {
	var filtered []model.FileEntry
	for _, file := range files {
		if !showHidden && file.IsHidden {
			continue
		}
		if f.shouldExclude(file.Name) {
			continue
		}
		if len(f.includePatterns) > 0 && !f.shouldInclude(file.Name) {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func (f *Filter) shouldExclude(name string) bool {
	for _, pattern := range f.excludePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func (f *Filter) shouldInclude(name string) bool {
	for _, pattern := range f.includePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
