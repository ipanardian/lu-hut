package updater

import (
	"strings"
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

func TestGetChecksumURL(t *testing.T) {
	release := &GitHubRelease{}
	release.Assets = append(release.Assets,
		struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{Name: "checksums.txt", BrowserDownloadURL: "https://example.test/checksums.txt"},
	)

	if got := getChecksumURL(release); got != "https://example.test/checksums.txt" {
		t.Fatalf("getChecksumURL() = %q, want checksum URL", got)
	}

	if got := getChecksumURL(&GitHubRelease{}); got != "" {
		t.Fatalf("getChecksumURL() = %q, want empty URL when asset is missing", got)
	}
}

func TestPerformUpdate_RejectsReleaseWithoutChecksums(t *testing.T) {
	release := &GitHubRelease{TagName: "v0.7.0"}
	release.Assets = append(release.Assets,
		struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{Name: GetArchiveName(release.TagName), BrowserDownloadURL: "http://127.0.0.1:1/archive.tar.gz"},
	)

	err := PerformUpdate(release)
	if err == nil {
		t.Fatal("PerformUpdate() error = nil, want missing checksum error")
	}
	if !strings.Contains(err.Error(), "does not contain checksums.txt") {
		t.Fatalf("PerformUpdate() error = %q, want missing checksum error", err)
	}
}

func TestParseChecksumLine(t *testing.T) {
	checksumContent := `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  lu-hut_0.6.0_Linux_x86_64.tar.gz
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9  lu-hut_0.6.0_Darwin_arm64.tar.gz
b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9  lu-hut_0.6.0_Darwin_amd64.tar.gz`

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
			expected:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:        "archive not found",
			archiveName: "lu-hut_0.6.0_Windows_amd64.tar.gz",
			shouldError: true,
		},
		{
			name:        "reject malformed checksum",
			archiveName: "lu-hut_0.6.0_Darwin_amd64.tar.gz",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := checksumContent
			if tt.name == "reject malformed checksum" {
				content = "not-a-sha256  " + tt.archiveName
			}

			result, err := parseChecksumFromContent(content, tt.archiveName)
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
