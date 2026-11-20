package lister

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ipanardian/lu-hut/internal/config"
)

func TestListPathsContinuesAfterInvalidPath(t *testing.T) {
	root := t.TempDir()
	validDir := filepath.Join(root, "valid")
	if err := os.Mkdir(validDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validDir, "file.txt"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		lst := New(config.NewDefaultConfig())
		err := lst.ListPaths([]string{filepath.Join(root, "missing"), validDir})
		if err == nil || !strings.Contains(err.Error(), "missing") {
			t.Fatalf("ListPaths() error = %v, want missing-path error", err)
		}
	})

	if !strings.Contains(output, "file.txt") {
		t.Fatalf("ListPaths() did not render valid path; output = %q", output)
	}
}

func TestListPathsSymlinks(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.Mkdir("dir", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"file.txt", "dir/child.txt"} {
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for name, target := range map[string]string{
		"file-link": "file.txt", "dir-link": "dir", "dangling": "missing",
	} {
		if err := os.Symlink(target, name); err != nil {
			t.Fatal(err)
		}
	}

	for _, tt := range []struct {
		arg, want, absent string
		wantErr           bool
	}{
		{arg: "file-link", want: "file-link -> file.txt"},
		{arg: "dir-link", want: "dir-link -> dir", absent: "child.txt"},
		{arg: "dir-link/", want: "child.txt", absent: "dir-link -> dir"},
		{arg: "dangling", want: "dangling -> missing"},
		{arg: "file-link/", wantErr: true},
		{arg: "missing", wantErr: true},
	} {
		t.Run(tt.arg, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.ColorMode, cfg.IconMode = "never", "never"
			var err error
			output := captureStdout(t, func() {
				err = New(cfg).ListPaths([]string{tt.arg})
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ListPaths() error = %v, want error: %t", err, tt.wantErr)
			}
			if !strings.Contains(output, tt.want) || (tt.absent != "" && strings.Contains(output, tt.absent)) {
				t.Errorf("ListPaths() output = %q, want %q and omit %q", output, tt.want, tt.absent)
			}
		})
	}
}

func TestListPathsFiltersFileOperands(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("src", 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"readme.txt", "src/main.go", "src/skip.go", "src/notes.txt"} {
		if err := os.WriteFile(name, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		name   string
		args   []string
		want   string
		absent []string
	}{
		{"files", []string{"readme.txt", "src/main.go", "src/skip.go"}, "src/main.go", []string{"readme.txt", "skip.go"}},
		{"reversed", []string{"src/main.go", "readme.txt"}, "src/main.go", []string{"readme.txt"}},
		{"mixed", []string{"src", "readme.txt"}, "main.go", []string{"readme.txt", "skip.go", "notes.txt"}},
		{"all filtered", []string{"readme.txt", "src/skip.go"}, "", []string{"readme.txt", "skip.go", "Name"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewDefaultConfig()
			cfg.ColorMode, cfg.IconMode = "never", "never"
			cfg.IncludePatterns = []string{"*.go"}
			cfg.ExcludePatterns = []string{"skip.go"}
			var err error
			output := captureStdout(t, func() { err = New(cfg).ListPaths(tt.args) })
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output, tt.want) {
				t.Errorf("ListPaths() output = %q, want %q", output, tt.want)
			}
			for _, name := range tt.absent {
				if strings.Contains(output, name) {
					t.Errorf("ListPaths() output = %q, unexpectedly contains %q", output, name)
				}
			}
		})
	}
}

func TestTreeIncludePrunesExcludedBranches(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, dir := range []string{"src/keep", "src/vendor", "src/excluded-only/vendor", "src/ignored-only"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.go"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile("src/.gitignore", []byte("ignored-only/main.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewDefaultConfig()
	cfg.Tree, cfg.GitIgnore = true, true
	cfg.ColorMode, cfg.IconMode = "never", "never"
	cfg.IncludePatterns = []string{"*.go"}
	cfg.ExcludePatterns = []string{"vendor"}
	var err error
	output := captureStdout(t, func() { err = New(cfg).ListPaths([]string{"src"}) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "keep/") || !strings.Contains(output, "main.go") {
		t.Errorf("tree output = %q, want keep/main.go", output)
	}
	for _, name := range []string{"vendor/", "excluded-only/", "ignored-only/"} {
		if strings.Contains(output, name) {
			t.Errorf("tree output = %q, unexpectedly contains %q", output, name)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer

	done := make(chan string)
	go func() {
		var buffer bytes.Buffer
		_, _ = io.Copy(&buffer, reader)
		done <- buffer.String()
	}()

	fn()
	_ = writer.Close()
	os.Stdout = original
	output := <-done
	_ = reader.Close()
	return output
}
