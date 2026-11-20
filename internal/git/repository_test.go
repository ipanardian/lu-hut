package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepositoryGetStatus(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T, root string, file string)
		status string
	}{
		{
			name: "untracked",
			setup: func(t *testing.T, root, file string) {
				writeFile(t, file, "new")
			},
			status: "?",
		},
		{
			name: "staged add",
			setup: func(t *testing.T, root, file string) {
				writeFile(t, file, "new")
				gitCommand(t, root, "add", "file.txt")
			},
			status: "A",
		},
		{
			name: "staged modification",
			setup: func(t *testing.T, root, file string) {
				writeFile(t, file, "before")
				gitCommand(t, root, "add", "file.txt")
				gitCommand(t, root, "-c", "user.name=Test User", "-c", "user.email=test@example.com", "commit", "-m", "initial")
				writeFile(t, file, "after")
				gitCommand(t, root, "add", "file.txt")
			},
			status: "M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			gitCommand(t, root, "init")
			file := filepath.Join(root, "file.txt")
			tt.setup(t, root, file)

			repo, err := NewRepository(root)
			if err != nil {
				t.Fatalf("NewRepository() error = %v", err)
			}
			if got := repo.GetStatus(file); got != tt.status {
				t.Errorf("GetStatus() = %q, want %q", got, tt.status)
			}
		})
	}
}

func gitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
