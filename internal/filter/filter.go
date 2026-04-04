// Package filter provides functionality for filtering file entries based on patterns.
package filter

import (
	"path/filepath"

	"github.com/ipanardian/lu-hut/internal/gitignore"
	"github.com/ipanardian/lu-hut/internal/model"
)

type Filter struct {
	includePatterns []string
	excludePatterns []string
	gitIgnore       *gitignore.Matcher
}

func NewFilter(includePatterns, excludePatterns []string, gitIgnore *gitignore.Matcher) *Filter {
	return &Filter{
		includePatterns: includePatterns,
		excludePatterns: excludePatterns,
		gitIgnore:       gitIgnore,
	}
}

func (f *Filter) Apply(files []model.FileEntry, showHidden bool, basePath string) []model.FileEntry {
	var filtered []model.FileEntry
	for _, file := range files {
		if !showHidden && file.IsHidden {
			continue
		}
		if f.gitIgnore != nil && file.Name == ".git" {
			continue
		}
		if f.shouldExclude(file.Name) {
			continue
		}
		if len(f.includePatterns) > 0 && !f.shouldInclude(file.Name) {
			continue
		}
		if f.gitIgnore != nil {
			filePath := filepath.Join(basePath, file.Name)
			if f.gitIgnore.IsIgnored(filePath, file.IsDir) {
				continue
			}
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

func (f *Filter) ShouldInclude(name string) bool {
	return f.shouldInclude(name)
}

func (f *Filter) ShouldExclude(name string) bool {
	return f.shouldExclude(name)
}

func (f *Filter) HasIncludePatterns() bool {
	return len(f.includePatterns) > 0
}

func (f *Filter) IsGitIgnored(path string, isDir bool) bool {
	if f.gitIgnore == nil {
		return false
	}
	if filepath.Base(path) == ".git" {
		return true
	}
	return f.gitIgnore.IsIgnored(path, isDir)
}
