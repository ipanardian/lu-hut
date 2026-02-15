// Package git provides Git repository integration for file status tracking.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repository struct {
	repoRoot     string
	statusCache  map[string]string
	statusLoaded bool
}

func NewRepository(path string) (*Repository, error) {
	root, err := findGitRoot(path)
	if err != nil {
		return nil, err
	}
	return &Repository{
		repoRoot:    root,
		statusCache: make(map[string]string),
	}, nil
}

func (g *Repository) loadAllStatus() error {
	if g.statusLoaded {
		return nil
	}

	cmd := exec.Command("git", "-C", g.repoRoot, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		staging := line[0]
		worktree := line[1]
		filePath := line[3:]

		if staging == 'R' || staging == 'C' {
			if idx := strings.Index(filePath, " -> "); idx != -1 {
				filePath = filePath[idx+4:]
			}
		}

		var status string
		if worktree != ' ' && worktree != '?' {
			switch worktree {
			case 'M':
				status = "M"
			case 'D':
				status = "D"
			case 'A':
				status = "A"
			case 'R':
				status = "R"
			case 'C':
				status = "C"
			}
		} else if staging != ' ' {
			switch staging {
			case 'A':
				status = "A"
			case 'M':
				status = "M"
			case 'D':
				status = "D"
			case 'R':
				status = "R"
			case 'C':
				status = "C"
			}
		} else if worktree == '?' {
			status = "?"
		}

		if status != "" {
			g.statusCache[filePath] = status
		}
	}

	g.statusLoaded = true
	return nil
}

func (g *Repository) GetStatus(filePath string) string {
	if err := g.loadAllStatus(); err != nil {
		return ""
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}

	relPath, err := filepath.Rel(g.repoRoot, absPath)
	if err != nil {
		return ""
	}

	relPath = filepath.ToSlash(relPath)

	if status, ok := g.statusCache[relPath]; ok {
		return status
	}

	return ""
}

func findGitRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}
