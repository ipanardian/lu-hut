// Package main provides tests for lu-hutg.
//
// Coordinate first, complain later.
//
// GitHub: https://github.com/ipanardian/lu-hutg
// Author: Ipan Ardian
// Version: v1.0.0
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculateDisplayWidths(t *testing.T) {
	tests := []struct {
		name     string
		data     [][]string
		expected []int
	}{
		{
			name: "basic table",
			data: [][]string{
				{"Name", "Size", "Modified"},
				{"file.txt", "1.2 KB", "2 minutes ago"},
				{"very-long-filename.go", "15.3 KB", "1 hour ago"},
			},
			expected: []int{21, 7, 13},
		},
		{
			name:     "empty table",
			data:     [][]string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widths := calculateDisplayWidths(tt.data)
			if len(widths) != len(tt.expected) {
				t.Errorf("got %d columns, want %d", len(widths), len(tt.expected))
				return
			}
			for i, w := range widths {
				if w != tt.expected[i] {
					t.Errorf("column %d: got width %d, want %d", i, w, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatPermissions(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected string
	}{
		{"regular file", 0644, "-rw-r--r--"},
		{"directory", 0755 | os.ModeDir, "drwxr-xr-x"},
		{"executable", 0755, "-rwxr-xr-x"},
		{"symlink", 0777 | os.ModeSymlink, "lrwxrwxrwx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPermissions(tt.mode)
			clean := stripANSI(result)
			if clean != tt.expected {
				t.Errorf("formatPermissions(%v) = %q, want %q", tt.mode, clean, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		isDir    bool
		expected string
	}{
		{"directory", 0, true, "-"},
		{"bytes", 512, false, "512 B"},
		{"kilobytes", 1536, false, "1.5 KB"},
		{"megabytes", 2097152, false, "2.0 MB"},
		{"gigabytes", 3221225472, false, "3.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size, tt.isDir)
			clean := stripANSI(result)
			if clean != tt.expected {
				t.Errorf("formatSize(%d, %v) = %q, want %q", tt.size, tt.isDir, clean, tt.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mRed Text\x1b[0m"
	expected := "Red Text"
	result := stripANSI(input)
	if result != expected {
		t.Errorf("stripANSI(%q) = %q, want %q", input, result, expected)
	}
}

func TestNameSortStrategy(t *testing.T) {
	now := time.Now()
	files := []FileEntry{
		{Name: "file.txt", ModTime: now.Add(-time.Hour), IsDir: false},
		{Name: "dir", ModTime: now.Add(-time.Minute), IsDir: true},
		{Name: "another.txt", ModTime: now, IsDir: false},
	}

	strategy := &NameSortStrategy{}
	strategy.Sort(files, false)

	if !files[0].IsDir {
		t.Error("first item should be a directory")
	}
	if files[0].Name != "dir" {
		t.Errorf("got %q, want 'dir'", files[0].Name)
	}
	if files[1].Name != "another.txt" {
		t.Errorf("got %q, want 'another.txt'", files[1].Name)
	}
}

func TestTimeSortStrategy(t *testing.T) {
	now := time.Now()
	files := []FileEntry{
		{Name: "old.txt", ModTime: now.Add(-time.Hour), IsDir: false},
		{Name: "new.txt", ModTime: now, IsDir: false},
		{Name: "medium.txt", ModTime: now.Add(-30 * time.Minute), IsDir: false},
	}

	strategy := &TimeSortStrategy{}
	strategy.Sort(files, false)

	if files[0].Name != "new.txt" {
		t.Errorf("got %q, want 'new.txt'", files[0].Name)
	}
	if files[2].Name != "old.txt" {
		t.Errorf("got %q, want 'old.txt'", files[2].Name)
	}
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()
	if width <= 0 {
		t.Error("terminal width should be positive")
	}
	if width > 1000 {
		t.Errorf("terminal width %d seems unrealistic", width)
	}
}

func TestLookupMin(t *testing.T) {
	tests := []struct {
		name     string
		mins     []int
		idx      int
		fallback int
		expected int
	}{
		{"within bounds", []int{10, 20, 30}, 1, 5, 20},
		{"out of bounds", []int{10, 20}, 5, 15, 15},
		{"zero value", []int{10, 0, 30}, 1, 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookupMin(tt.mins, tt.idx, tt.fallback)
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFileFilter(t *testing.T) {
	files := []FileEntry{
		{Name: "main.go"},
		{Name: "main_test.go"},
		{Name: "README.md"},
		{Name: "temp.tmp"},
	}

	t.Run("include pattern", func(t *testing.T) {
		filter := &FileFilter{includePatterns: []string{"*.go"}}
		filtered := filter.Apply(files, true)
		if len(filtered) != 2 {
			t.Errorf("got %d files, want 2", len(filtered))
		}
	})

	t.Run("exclude pattern", func(t *testing.T) {
		filter := &FileFilter{excludePatterns: []string{"*.tmp"}}
		filtered := filter.Apply(files, true)
		if len(filtered) != 3 {
			t.Errorf("got %d files, want 3", len(filtered))
		}
	})

	t.Run("include and exclude", func(t *testing.T) {
		filter := &FileFilter{
			includePatterns: []string{"*.go"},
			excludePatterns: []string{"*_test.go"},
		}
		filtered := filter.Apply(files, true)
		if len(filtered) != 1 {
			t.Errorf("got %d files, want 1", len(filtered))
		}
		if filtered[0].Name != "main.go" {
			t.Errorf("got %q, want 'main.go'", filtered[0].Name)
		}
	})
}

func TestDirectoryLister(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	err := os.WriteFile(tmpFile, content, 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	config := Config{}
	lister := NewDirectoryLister(config)
	err = lister.List(tmpDir)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
}

func TestFindGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	root, err := findGitRoot(tmpDir)
	if err != nil {
		t.Errorf("findGitRoot() error = %v", err)
	}
	if root != tmpDir {
		t.Errorf("got %q, want %q", root, tmpDir)
	}
}

func TestDirectoryListerRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure:
	// tmpDir/
	//   file1.txt
	//   subdir1/
	//     file2.txt
	//     subdir2/
	//       file3.txt
	//   subdir3/
	//     file4.txt

	// Create files and directories
	files := map[string]string{
		"file1.txt":                 "content1",
		"subdir1/file2.txt":         "content2",
		"subdir1/subdir2/file3.txt": "content3",
		"subdir3/file4.txt":         "content4",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create file %s: %v", fullPath, err)
		}
	}

	t.Run("recursive listing", func(t *testing.T) {
		config := Config{Recursive: true}
		lister := NewDirectoryLister(config)
		err := lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})

	t.Run("recursive with hidden files", func(t *testing.T) {
		// Create a hidden directory
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		err := os.Mkdir(hiddenDir, 0755)
		if err != nil {
			t.Fatalf("failed to create hidden directory: %v", err)
		}
		err = os.WriteFile(filepath.Join(hiddenDir, "hidden.txt"), []byte("hidden"), 0644)
		if err != nil {
			t.Fatalf("failed to create hidden file: %v", err)
		}

		config := Config{Recursive: true, ShowHidden: true}
		lister := NewDirectoryLister(config)
		err = lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})

	t.Run("recursive with filter", func(t *testing.T) {
		config := Config{
			Recursive:       true,
			IncludePatterns: []string{"*.txt"},
		}
		lister := NewDirectoryLister(config)
		err := lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})
}

func TestDirectoryListerRecursiveDepthLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deep directory structure
	deepPath := tmpDir
	for i := 0; i < 10; i++ {
		deepPath = filepath.Join(deepPath, fmt.Sprintf("level%d", i))
		err := os.MkdirAll(deepPath, 0755)
		if err != nil {
			t.Fatalf("failed to create deep directory: %v", err)
		}
		err = os.WriteFile(filepath.Join(deepPath, "file.txt"), []byte("content"), 0644)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	t.Run("depth limit of 3", func(t *testing.T) {
		config := Config{
			Recursive: true,
			MaxDepth:  3,
		}
		lister := NewDirectoryLister(config)
		err := lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})

	t.Run("depth limit of 0 uses default", func(t *testing.T) {
		config := Config{
			Recursive: true,
			MaxDepth:  0,
		}
		lister := NewDirectoryLister(config)
		err := lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})

	t.Run("depth limit of 1", func(t *testing.T) {
		config := Config{
			Recursive: true,
			MaxDepth:  1,
		}
		lister := NewDirectoryLister(config)
		err := lister.List(tmpDir)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
	})
}
