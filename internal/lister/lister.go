// Package lister provides directory listing functionality.
package lister

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/filter"
	"github.com/ipanardian/lu-hut/internal/git"
	"github.com/ipanardian/lu-hut/internal/gitignore"
	"github.com/ipanardian/lu-hut/internal/model"
	"github.com/ipanardian/lu-hut/internal/renderer"
	"github.com/ipanardian/lu-hut/internal/sort"
)

type Lister struct {
	config    config.Config
	gitRepo   *git.Repository
	gitIgnore *gitignore.Matcher
	filter    *filter.Filter
	sortStrat sort.Strategy
}

func New(cfg config.Config) *Lister {
	switch cfg.ColorMode {
	case "never":
		color.NoColor = true
	case "always":
		color.NoColor = false
	case "auto":
	}

	filter := filter.NewFilter(cfg.IncludePatterns, cfg.ExcludePatterns, nil)

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

	return &Lister{
		config:    cfg,
		filter:    filter,
		sortStrat: sortStrat,
	}
}

func (d *Lister) List(path string) error {
	return d.ListPaths([]string{path})
}

func (d *Lister) ListPaths(paths []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	var filePaths []string
	var dirPaths []string
	var pathErrors []error

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			pathErrors = append(pathErrors, fmt.Errorf("cannot resolve '%s': %w", p, err))
			continue
		}
		info, err := os.Stat(absPath)
		if err != nil {
			pathErrors = append(pathErrors, fmt.Errorf("cannot access '%s': %w", p, err))
			continue
		}
		if info.IsDir() {
			dirPaths = append(dirPaths, absPath)
		} else {
			filePaths = append(filePaths, absPath)
		}
	}

	multipleSources := len(filePaths)+len(dirPaths) > 1

	if len(filePaths) > 0 {
		files := d.collectFileEntries(filePaths)
		d.sortStrat.Sort(files, d.config.Reverse)
		tableRenderer := renderer.NewTable(d.config)
		tableRenderer.Render(files, time.Now())
	}

	for i, dirPath := range dirPaths {
		if multipleSources {
			if i > 0 || len(filePaths) > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", color.New(color.FgCyan, color.Bold).Sprint(dirPath))
		}
		if err := d.listSingle(ctx, dirPath); err != nil {
			pathErrors = append(pathErrors, fmt.Errorf("cannot list '%s': %w", dirPath, err))
		}
	}

	return errors.Join(pathErrors...)
}

func (d *Lister) listSingle(ctx context.Context, absPath string) error {
	if d.config.ShowGit {
		d.gitRepo, _ = git.NewRepository(absPath)
	}

	if d.config.GitIgnore {
		d.gitIgnore, _ = gitignore.NewMatcher(absPath)
		d.filter = filter.NewFilter(d.config.IncludePatterns, d.config.ExcludePatterns, d.gitIgnore)
	}

	if d.config.Tree {
		return d.listTree(ctx, absPath)
	}

	if d.config.Recursive {
		return d.listRecursive(ctx, absPath)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	files := d.collectFiles(absPath, entries)
	files = d.filter.Apply(files, d.config.ShowHidden, absPath)
	d.sortStrat.Sort(files, d.config.Reverse)

	tableRenderer := renderer.NewTable(d.config)
	tableRenderer.Render(files, time.Now())

	return nil
}

func (d *Lister) listTree(ctx context.Context, rootPath string) error {
	treeRenderer := renderer.NewTree(d.config)
	if d.gitRepo != nil {
		treeRenderer.SetGitRepo(d.gitRepo)
	}
	treeRenderer.SetFilter(d.filter)
	return treeRenderer.Render(ctx, rootPath, time.Now())
}

func (d *Lister) listRecursive(ctx context.Context, rootPath string) error {
	var (
		maxDepth = d.config.MaxDepth
		maxDirs  = 10000
	)
	type dirEntry struct {
		path  string
		level int
	}

	dirs := []dirEntry{{path: rootPath, level: 0}}
	dirCount := 0

	for len(dirs) > 0 {
		select {
		case <-ctx.Done():
			fmt.Println("\nOperation cancelled by user")
			return ctx.Err()
		default:
		}

		current := dirs[0]
		dirs = dirs[1:]

		if maxDepth > 0 && current.level >= maxDepth {
			if current.level == maxDepth {
				indent := ""
				if current.level > 0 {
					indent = strings.Repeat("  ", current.level-1)
				}
				fmt.Printf("\n%s%s: (max depth reached)\n", indent, current.path)
			}
			continue
		}

		dirCount++
		if dirCount > maxDirs {
			fmt.Printf("\nReached maximum directory limit (%d). Stopping recursion.\n", maxDirs)
			break
		}

		if current.level > 0 {
			indent := strings.Repeat("  ", current.level-1)
			fmt.Printf("\n%s%s:\n", indent, current.path)
		}

		entries, err := os.ReadDir(current.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", current.path, err)
			continue
		}

		files := d.collectFiles(current.path, entries)
		files = d.filter.Apply(files, d.config.ShowHidden, current.path)
		d.sortStrat.Sort(files, d.config.Reverse)

		if len(files) == 0 {
			continue
		}

		renderer := renderer.NewTable(d.config)
		renderer.Render(files, time.Now())

		for _, file := range files {
			if file.IsDir {
				nextLevel := current.level + 1
				if maxDepth > 0 && nextLevel >= maxDepth {
					continue
				}
				dirPath := filepath.Join(current.path, file.Name)
				dirs = append(dirs, dirEntry{path: dirPath, level: nextLevel})
			}
		}
	}

	return nil
}

func (d *Lister) collectFileEntries(paths []string) []model.FileEntry {
	cwd, _ := os.Getwd()
	gitRepos := make(map[string]*git.Repository)

	var files []model.FileEntry
	for _, absPath := range paths {
		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot stat '%s': %v\n", absPath, err)
			continue
		}

		parentDir := filepath.Dir(absPath)
		relPath, _ := filepath.Rel(cwd, absPath)
		if relPath == "" {
			relPath = absPath
		}

		file := model.FileEntry{
			Name:     relPath,
			Path:     absPath,
			Size:     info.Size(),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
			IsDir:    info.IsDir(),
			IsHidden: strings.HasPrefix(filepath.Base(absPath), "."),
		}

		if d.config.ShowGit && !file.IsDir {
			if repo, ok := gitRepos[parentDir]; ok {
				file.GitStatus = repo.GetStatus(absPath)
			} else if repo, err := git.NewRepository(parentDir); err == nil {
				gitRepos[parentDir] = repo
				file.GitStatus = repo.GetStatus(absPath)
			}
		}

		if d.config.ShowUser {
			file.Author, file.Group = extractUserGroup(info)
		}

		files = append(files, file)
	}

	return files
}

func (d *Lister) collectFiles(path string, entries []fs.DirEntry) []model.FileEntry {
	files := make([]model.FileEntry, 0, len(entries))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s: %v\n", entry.Name(), err)
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

		if d.config.ShowGit && d.gitRepo != nil && !file.IsDir {
			file.GitStatus = d.gitRepo.GetStatus(file.Path)
		}

		if d.config.ShowUser {
			file.Author, file.Group = extractUserGroup(info)
		}

		files = append(files, file)
	}

	return files
}
