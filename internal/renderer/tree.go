// Package renderer provides tree rendering functionality.
package renderer

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/filter"
	"github.com/ipanardian/lu-hut/internal/git"
	"github.com/ipanardian/lu-hut/internal/icons"
	"github.com/ipanardian/lu-hut/internal/model"
	"github.com/ipanardian/lu-hut/internal/sort"
	"github.com/ipanardian/lu-hut/pkg/helper"
)

type Tree struct {
	config       config.Config
	gitRepo      *git.Repository
	sortStrategy sort.Strategy
	filter       *filter.Filter
	maxPermWidth int
	maxSizeWidth int
	maxModWidth  int
	maxUserWidth int
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

	if r.config.ShowLong || r.config.ShowUser {
		if err := r.calculateColumnWidths(ctx, path, 0, now); err != nil {
			return err
		}
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

		if r.config.ShowLong || r.config.ShowUser {
			file.Author, file.Group = r.extractUserGroup(info)
		}

		if r.config.ShowGit && r.gitRepo != nil {
			file.GitStatus = r.gitRepo.GetStatus(file.Path)
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
				if r.filter.IsGitIgnored(file.Path, file.IsDir) {
					continue
				}
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
				if r.filter.ShouldExclude(file.Name) {
					continue
				}
				if r.filter.IsGitIgnored(file.Path, file.IsDir) {
					continue
				}
				filtered = append(filtered, file)
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

		var line string
		if r.config.ShowLong || r.config.ShowUser {
			metadataPrefix := r.formatMetadataPrefixAligned(file, now)
			line = metadataPrefix + prefix + connector
		} else {
			line = prefix + connector
		}

		iconMode := icons.Mode(r.config.IconMode)

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
			line += formatName(file, dirWidth, iconMode) + "/"
		} else {
			line += formatName(file, nameWidth, iconMode)
		}

		if r.config.ShowGit && r.gitRepo != nil && !file.IsDir {
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
			if err := r.renderTreeRecursive(ctx, file.Path, newPrefix, true, level+1, now); err != nil {
				continue
			}
		}
	}

	return nil
}

func (r *Tree) hasMatchingDescendants(ctx context.Context, dirPath string) bool {
	var result bool

	if err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
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
	}); err != nil {
		return false
	}

	if ctx.Err() != nil {
		return false
	}

	return result
}

func (r *Tree) extractUserGroup(fileInfo os.FileInfo) (string, string) {
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		u, errU := user.LookupId(strconv.Itoa(int(stat.Uid)))
		g, errG := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

		username := "unknown"
		groupname := "unknown"

		if errU == nil {
			username = u.Username
		}
		if errG == nil {
			groupname = g.Name
		}

		return username, groupname
	}
	return "unknown", "unknown"
}

func (r *Tree) calculateColumnWidths(ctx context.Context, path string, level int, now time.Time) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if r.config.MaxDepth > 0 && level >= r.config.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !r.config.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := model.FileEntry{
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		}

		if r.config.ShowLong || r.config.ShowUser {
			file.Author, _ = r.extractUserGroup(info)
		}

		perms := helper.StripANSI(formatPermissions(file.Mode, r.config.ShowOctal))
		size := helper.StripANSI(formatSize(file.Size, file.IsDir))
		modified := helper.StripANSI(formatModified(file.ModTime, now, r.config.ShowExactTime))
		user := file.Author
		if user == "" {
			user = "?"
		}

		if len(perms) > r.maxPermWidth {
			r.maxPermWidth = len(perms)
		}
		if len(size) > r.maxSizeWidth {
			r.maxSizeWidth = len(size)
		}
		if len(modified) > r.maxModWidth {
			r.maxModWidth = len(modified)
		}
		if len(user) > r.maxUserWidth {
			r.maxUserWidth = len(user)
		}

		if file.IsDir {
			if err := r.calculateColumnWidths(ctx, file.Path, level+1, now); err != nil {
				continue
			}
		}
	}

	return nil
}

func (r *Tree) formatMetadataPrefixAligned(file model.FileEntry, now time.Time) string {
	if r.config.ShowLong {
		return r.formatLongMetadata(file, now)
	} else if r.config.ShowUser {
		return r.formatUserMetadata(file)
	}
	return ""
}

func (r *Tree) formatLongMetadata(file model.FileEntry, now time.Time) string {
	perms := formatPermissions(file.Mode, r.config.ShowOctal)
	size := formatSize(file.Size, file.IsDir)

	permsStr := r.padRight(perms, r.maxPermWidth)
	sizeStr := r.padLeft(size, r.maxSizeWidth)

	termWidth := getTerminalWidth()
	if termWidth < 80 {
		return permsStr + " " + sizeStr + " "
	}

	modified := formatModified(file.ModTime, now, r.config.ShowExactTime)
	modifiedStr := r.padRight(modified, r.maxModWidth)

	user := file.Author
	if user == "" {
		user = "?"
	}
	userStr := r.padRight(user, r.maxUserWidth)

	return permsStr + " " + sizeStr + " " + modifiedStr + " " + userStr + " "
}

func (r *Tree) formatUserMetadata(file model.FileEntry) string {
	user := file.Author
	if user == "" {
		user = "?"
	}
	return r.padRight(user, r.maxUserWidth) + " "
}

func (r *Tree) padRight(s string, maxWidth int) string {
	displayWidth := len(helper.StripANSI(s))
	var padding strings.Builder
	for i := displayWidth; i < maxWidth; i++ {
		padding.WriteString(" ")
	}
	return s + padding.String()
}

func (r *Tree) padLeft(s string, maxWidth int) string {
	displayWidth := len(helper.StripANSI(s))
	var padding strings.Builder
	for i := displayWidth; i < maxWidth; i++ {
		padding.WriteString(" ")
	}
	return padding.String() + s
}
