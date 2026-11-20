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
