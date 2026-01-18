package renderer

import (
	"io/fs"
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
		mode     fs.FileMode
		expected string
	}{
		{
			name:     "directory",
			mode:     fs.ModeDir | 0o755,
			expected: "drwxr-xr-x",
		},
		{
			name:     "regular file",
			mode:     0o644,
			expected: "-rw-r--r--",
		},
		{
			name:     "executable",
			mode:     0o755,
			expected: "-rwxr-xr-x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPermissions(tt.mode)
			if result != tt.expected {
				t.Errorf("formatPermissions(%o) = %q, want %q", tt.mode, result, tt.expected)
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
		{
			name:     "directory",
			size:     0,
			isDir:    true,
			expected: "-",
		},
		{
			name:     "bytes",
			size:     512,
			isDir:    false,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			size:     1024,
			isDir:    false,
			expected: "1.0 KB",
		},
		{
			name:     "megabytes",
			size:     1024 * 1024,
			isDir:    false,
			expected: "1.0 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.size, tt.isDir)
			if result != tt.expected {
				t.Errorf("formatSize(%d, %v) = %q, want %q", tt.size, tt.isDir, result, tt.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "with ANSI codes",
			input:    "\x1b[31mred\x1b[0m text",
			expected: "red text",
		},
		{
			name:     "multiple ANSI codes",
			input:    "\x1b[1m\x1b[32mbold green\x1b[0m",
			expected: "bold green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()
	if width <= 0 {
		t.Errorf("getTerminalWidth() returned %d, want positive value", width)
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
		{
			name:     "index exists",
			mins:     []int{10, 20, 30},
			idx:      1,
			fallback: 5,
			expected: 20,
		},
		{
			name:     "index out of bounds",
			mins:     []int{10, 20, 30},
			idx:      5,
			fallback: 5,
			expected: 5,
		},
		{
			name:     "empty mins",
			mins:     []int{},
			idx:      0,
			fallback: 5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lookupMin(tt.mins, tt.idx, tt.fallback)
			if result != tt.expected {
				t.Errorf("lookupMin(%v, %d, %d) = %d, want %d", tt.mins, tt.idx, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestFormatModified(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name      string
		t         time.Time
		showExact bool
		expected  string
	}{
		{
			name:      "relative - seconds ago",
			t:         now.Add(-30 * time.Second),
			showExact: false,
			expected:  "30 seconds ago",
		},
		{
			name:      "relative - minutes ago",
			t:         now.Add(-5 * time.Minute),
			showExact: false,
			expected:  "5 minutes ago",
		},
		{
			name:      "relative - hours ago",
			t:         now.Add(-2 * time.Hour),
			showExact: false,
			expected:  "2 hours ago",
		},
		{
			name:      "exact time",
			t:         time.Date(2024, 1, 15, 9, 30, 0, 0, time.UTC),
			showExact: true,
			expected:  "2024-01-15 09:30:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatModified(tt.t, now, tt.showExact)
			if result != tt.expected {
				t.Errorf("formatModified(%v, %v, %v) = %q, want %q", tt.t, now, tt.showExact, result, tt.expected)
			}
		})
	}
}
