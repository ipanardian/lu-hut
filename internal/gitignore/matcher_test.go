package gitignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewMatcher(t *testing.T) {
	t.Run("no gitignore file", func(t *testing.T) {
		tmpDir := t.TempDir()
		matcher, err := NewMatcher(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if matcher == nil {
			t.Fatal("expected matcher to not be nil")
		}

		if matcher.IsIgnored(filepath.Join(tmpDir, "test.txt")) {
			t.Error("expected file not to be ignored when no .gitignore exists")
		}
	})

	t.Run("with gitignore file", func(t *testing.T) {
		tmpDir := t.TempDir()

		if err := os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755); err != nil {
			t.Fatalf("failed to create node_modules: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(tmpDir, "build"), 0755); err != nil {
			t.Fatalf("failed to create build: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(tmpDir, "src"), 0755); err != nil {
			t.Fatalf("failed to create src: %v", err)
		}

		gitignoreContent := `*.log
node_modules/
build/
*.tmp
`
		if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
			t.Fatalf("failed to create .gitignore: %v", err)
		}

		matcher, err := NewMatcher(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tests := []struct {
			path     string
			isDir    bool
			expected bool
		}{
			{filepath.Join(tmpDir, "test.log"), false, true},
			{filepath.Join(tmpDir, "debug.log"), false, true},
			{filepath.Join(tmpDir, "file.tmp"), false, true},
			{filepath.Join(tmpDir, "main.go"), false, false},
			{filepath.Join(tmpDir, "README.md"), false, false},
			{filepath.Join(tmpDir, "node_modules"), true, true},
			{filepath.Join(tmpDir, "build"), true, true},
			{filepath.Join(tmpDir, "src"), true, false},
		}

		for _, tc := range tests {
			result := matcher.IsIgnored(tc.path, tc.isDir)
			if result != tc.expected {
				t.Errorf("IsIgnored(%q, isDir=%v) = %v, want %v", tc.path, tc.isDir, result, tc.expected)
			}
		}
	})

	t.Run("gitignore with comments and empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()

		if err := os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755); err != nil {
			t.Fatalf("failed to create node_modules: %v", err)
		}

		gitignoreContent := `# This is a comment
*.log

# Another comment
node_modules/

`
		if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
			t.Fatalf("failed to create .gitignore: %v", err)
		}

		matcher, err := NewMatcher(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !matcher.IsIgnored(filepath.Join(tmpDir, "test.log"), false) {
			t.Error("expected .log files to be ignored")
		}

		if !matcher.IsIgnored(filepath.Join(tmpDir, "node_modules"), true) {
			t.Error("expected node_modules to be ignored")
		}
	})

	t.Run("negation patterns", func(t *testing.T) {
		tmpDir := t.TempDir()

		gitignoreContent := `*.log
!important.log
`
		if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
			t.Fatalf("failed to create .gitignore: %v", err)
		}

		matcher, err := NewMatcher(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !matcher.IsIgnored(filepath.Join(tmpDir, "debug.log"), false) {
			t.Error("expected debug.log to be ignored")
		}

		if matcher.IsIgnored(filepath.Join(tmpDir, "important.log"), false) {
			t.Error("expected important.log to NOT be ignored (negation)")
		}
	})
}

func TestMatcherNil(t *testing.T) {
	var m *Matcher
	if m.IsIgnored("/any/path") {
		t.Error("nil matcher should not ignore any files")
	}
}
