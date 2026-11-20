package filter

import (
	"testing"

	"github.com/ipanardian/lu-hut/internal/model"
)

func TestFileFilter(t *testing.T) {
	files := []model.FileEntry{
		{Name: "batu-bara.txt", IsHidden: false},
		{Name: "sawit.hidden", IsHidden: true},
		{Name: "lord.go", IsHidden: false},
		{Name: "semua-saya-yang-urus.go", IsHidden: false},
		{Name: "tambang.py", IsHidden: false},
	}

	t.Run("show hidden false", func(t *testing.T) {
		filter := NewFilter(nil, nil, nil)
		result := filter.Apply(files, false, "/tmp")

		if len(result) != 4 {
			t.Errorf("expected 4 files, got %d", len(result))
		}

		for _, f := range result {
			if f.IsHidden {
				t.Errorf("hidden file %s should not be included", f.Name)
			}
		}
	})

	t.Run("show hidden true", func(t *testing.T) {
		filter := NewFilter(nil, nil, nil)
		result := filter.Apply(files, true, "/tmp")

		if len(result) != 5 {
			t.Errorf("expected 5 files, got %d", len(result))
		}
	})

	t.Run("include pattern", func(t *testing.T) {
		filter := NewFilter([]string{"lord.go"}, nil, nil)
		result := filter.Apply(files, false, "/tmp")

		if len(result) != 1 {
			t.Errorf("expected 1 file, got %d", len(result))
		}

		if result[0].Name != "lord.go" {
			t.Errorf("expected lord.go, got %s", result[0].Name)
		}
	})

	t.Run("exclude pattern", func(t *testing.T) {
		filter := NewFilter(nil, []string{"*tambang.py"}, nil)
		result := filter.Apply(files, false, "/tmp")

		if len(result) != 3 {
			t.Errorf("expected 3 files, got %d", len(result))
		}

		for _, f := range result {
			if f.Name == "tambang.py" {
				t.Errorf("tambang.py should be excluded")
			}
		}
	})

	t.Run("include and exclude", func(t *testing.T) {
		filter := NewFilter([]string{"*.py"}, []string{"*.go"}, nil)
		result := filter.Apply(files, false, "/tmp")

		if len(result) != 1 {
			t.Errorf("expected 1 file, got %d", len(result))
		}

		if result[0].Name != "tambang.py" {
			t.Errorf("expected tambang.py, got %s", result[0].Name)
		}
	})
}

func TestShouldTraverseDir(t *testing.T) {
	t.Run("hidden dir excluded when showHidden false", func(t *testing.T) {
		f := NewFilter(nil, nil, nil)
		if f.ShouldTraverseDir(".git", true, false, "/tmp/.git") {
			t.Error("expected hidden dir to not be traversed")
		}
	})

	t.Run("hidden dir included when showHidden true", func(t *testing.T) {
		f := NewFilter(nil, nil, nil)
		if !f.ShouldTraverseDir(".config", true, true, "/tmp/.config") {
			t.Error("expected hidden dir to be traversed when showHidden is true")
		}
	})

	t.Run(".git dir always excluded", func(t *testing.T) {
		f := NewFilter(nil, nil, nil)
		if f.ShouldTraverseDir(".git", true, true, "/tmp/.git") {
			t.Error("expected .git dir to not be traversed even with showHidden true")
		}
	})

	t.Run("excluded dir not traversed", func(t *testing.T) {
		f := NewFilter(nil, []string{"vendor"}, nil)
		if f.ShouldTraverseDir("vendor", false, false, "/tmp/vendor") {
			t.Error("expected excluded dir to not be traversed")
		}
	})

	t.Run("include pattern does not prevent traversal", func(t *testing.T) {
		f := NewFilter([]string{"*.go"}, nil, nil)
		if !f.ShouldTraverseDir("internal", false, false, "/tmp/internal") {
			t.Error("expected dir to be traversed regardless of include pattern")
		}
	})

	t.Run("normal dir traversed", func(t *testing.T) {
		f := NewFilter([]string{"*.go"}, nil, nil)
		if !f.ShouldTraverseDir("src", false, false, "/tmp/src") {
			t.Error("expected normal dir to be traversed")
		}
	})
}
