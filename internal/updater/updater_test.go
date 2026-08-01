package updater

import (
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world",
			data:     []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSHA256(tt.data)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseChecksumLine(t *testing.T) {
	checksumContent := `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  lu-hut_0.6.0_Linux_x86_64.tar.gz
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9  lu-hut_0.6.0_Darwin_arm64.tar.gz
abc123def456  lu-hut_0.6.0_Darwin_amd64.tar.gz`

	tests := []struct {
		name        string
		archiveName string
		expected    string
		shouldError bool
	}{
		{
			name:        "find Linux archive",
			archiveName: "lu-hut_0.6.0_Linux_x86_64.tar.gz",
			expected:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:        "find Darwin arm64 archive",
			archiveName: "lu-hut_0.6.0_Darwin_arm64.tar.gz",
			expected:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:        "find Darwin amd64 archive",
			archiveName: "lu-hut_0.6.0_Darwin_amd64.tar.gz",
			expected:    "abc123def456",
		},
		{
			name:        "archive not found",
			archiveName: "lu-hut_0.6.0_Windows_amd64.tar.gz",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseChecksumFromContent(checksumContent, tt.archiveName)
			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
