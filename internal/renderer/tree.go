// Package renderer provides tree rendering functionality.
package renderer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/filter"
	"github.com/ipanardian/lu-hut/internal/git"
	"github.com/ipanardian/lu-hut/internal/model"
	"github.com/ipanardian/lu-hut/internal/sort"
	"github.com/ipanardian/lu-hut/pkg/helper"
)

type Tree struct {
	config       config.Config
	gitRepo      *git.Repository
	sortStrategy sort.Strategy
	filter       *filter.Filter
}

func NewTree(cfg config.Config) *Tree {
	var sortStrat sort.Strategy
	if cfg.SortSize {
		sortStrat = &sort.Size{}
	} else if cfg.SortExtension {
		sortStrat = &sort.Extension{}
	} else if cfg.SortModified {
		sortStrat = &sort.Time{}
	} else {
		sortStrat = &sort.Name{}
	}

	return &Tree{
		config:       cfg,
		sortStrategy: sortStrat,
	}
}

func (r *Tree) SetGitRepo(repo *git.Repository) {
	r.gitRepo = repo
}

func (r *Tree) SetFilter(f *filter.Filter) {
	r.filter = f
}

func (r *Tree) Render(ctx context.Context, path string, now time.Time) error {
	if ctx == nil {
		ctx = context.Background()
	}

	err := r.renderTreeRecursive(ctx, path, "", true, 0, now)
	if err == context.Canceled {
		fmt.Println("\nOperation cancelled by user")
		err = nil
	}
	return err
}

func (r *Tree) renderTreeRecursive(ctx context.Context, path string, prefix string, _ bool, level int, now time.Time) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if r.config.MaxDepth > 0 && level >= r.config.MaxDepth {
		if level == r.config.MaxDepth {
			fmt.Printf("%s└── (max depth reached)\n", prefix)
		}
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("%s├── Error: %v\n", prefix, err)
		return nil
	}

	files := make([]model.FileEntry, 0, len(entries))
	for _, entry := range entries {
		if !r.config.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := model.FileEntry{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Size:     info.Size(),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
			IsDir:    entry.IsDir(),
			IsHidden: strings.HasPrefix(entry.Name(), "."),
		}

		files = append(files, file)
	}

	if r.sortStrategy != nil {
		r.sortStrategy.Sort(files, r.config.Reverse)
	}

	if r.filter != nil {
		if r.filter.HasIncludePatterns() {
			var filtered []model.FileEntry

			for _, file := range files {
				if file.IsDir {
					if r.hasMatchingDescendants(ctx, file.Path) {
						filtered = append(filtered, file)
					}
				} else {
					if r.filter.ShouldInclude(file.Name) && !r.filter.ShouldExclude(file.Name) {
						filtered = append(filtered, file)
					}
				}
			}

			files = filtered
		} else {
			var filtered []model.FileEntry
			for _, file := range files {
				if !r.filter.ShouldExclude(file.Name) {
					filtered = append(filtered, file)
				}
			}
			files = filtered
		}
	}

	if r.sortStrategy != nil {
		r.sortStrategy.Sort(files, r.config.Reverse)
	}

	for i, file := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		isLast := i == len(files)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		line := prefix + connector
		nameWidth := getTerminalWidth()
		if nameWidth <= 0 {
			nameWidth = defaultNameMaxWidth
		}
		prefixWidth := runeCount(helper.StripANSI(line))
		nameWidth -= prefixWidth
		if nameWidth <= 0 {
			nameWidth = defaultNameMaxWidth
		}

		if file.IsDir {
			dirWidth := nameWidth
			if dirWidth > 1 {
				dirWidth--
			}
			line += formatName(file, dirWidth) + "/"
		} else {
			line += formatName(file, nameWidth)
		}

		if r.config.ShowGit && r.gitRepo != nil {
			if status := r.gitRepo.GetStatus(file.Path); status != "" {
				line += " " + formatGitStatus(status)
			}
		}

		fmt.Println(line)

		if file.IsDir {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			r.renderTreeRecursive(ctx, file.Path, newPrefix, true, level+1, now)
		}
	}

	return nil
}

func (r *Tree) hasMatchingDescendants(ctx context.Context, dirPath string) bool {
	var result bool

	filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			return nil
		}

		if strings.Count(path, string(filepath.Separator))-strings.Count(dirPath, string(filepath.Separator)) > 5 {
			return filepath.SkipDir
		}

		if !r.config.ShowHidden && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			if r.filter.ShouldInclude(d.Name()) && !r.filter.ShouldExclude(d.Name()) {
				result = true
				return filepath.SkipAll
			}
		}

		return nil
	})

	if ctx.Err() != nil {
		return false
	}

	return result
}
