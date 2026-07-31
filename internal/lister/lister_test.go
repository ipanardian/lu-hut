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
