package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameWithReplaceOverwritesExistingDestination(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	if err := os.WriteFile(src, []byte("new-content"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old-content"), 0644); err != nil {
		t.Fatalf("write dst: %v", err)
	}

	if err := renameWithReplace(src, dst); err != nil {
		t.Fatalf("renameWithReplace failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "new-content" {
		t.Errorf("dst content = %q, want %q", data, "new-content")
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src should no longer exist after rename")
	}
}

func TestRenameWithReplaceCreatesNewDestination(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := renameWithReplace(src, dst); err != nil {
		t.Fatalf("renameWithReplace failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("dst content = %q, want %q", data, "content")
	}
}
