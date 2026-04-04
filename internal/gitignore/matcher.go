// Package gitignore provides functionality for parsing and matching .gitignore patterns.
package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Matcher struct {
	matcher gitignore.Matcher
	root    string
}

func NewMatcher(dir string) (*Matcher, error) {
	absRoot, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	patterns, err := loadGitignore(absRoot)
	if err != nil {
		return nil, err
	}

	return &Matcher{
		matcher: gitignore.NewMatcher(patterns),
		root:    absRoot,
	}, nil
}

func (m *Matcher) IsIgnored(path string, isDir ...bool) bool {
	if m == nil {
		return false
	}

	relPath, err := filepath.Rel(m.root, path)
	if err != nil {
		return false
	}

	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	isDirectory := false
	if len(isDir) > 0 {
		isDirectory = isDir[0]
	} else {
		info, err := os.Stat(path)
		isDirectory = err == nil && info.IsDir()
	}

	return m.matcher.Match(parts, isDirectory)
}

func loadGitignore(root string) ([]gitignore.Pattern, error) {
	gitignorePath := filepath.Join(root, ".gitignore")

	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
