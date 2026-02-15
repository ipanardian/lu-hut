// Package updater provides version checking and self-update functionality for lu-hut.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ipanardian/lu-hut/internal/constants"
)

const (
	githubAPIURL = "https://api.github.com/repos/ipanardian/lu-hut/releases/latest"
	checkTimeout = 10 * time.Second
)

type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestVersion() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: checkTimeout,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func IsNewerVersion(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == latest {
		return false
	}

	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	maxLen := max(len(currentParts), len(latestParts))

	for i := 0; i < maxLen; i++ {
		var currentNum, latestNum int

		if i < len(currentParts) {
			_, _ = fmt.Sscanf(currentParts[i], "%d", &currentNum)
		}
		if i < len(latestParts) {
			_, _ = fmt.Sscanf(latestParts[i], "%d", &latestNum)
		}

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	return false
}

func GetCurrentVersion() string {
	return constants.Version
}

func GetBinaryName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	return fmt.Sprintf("lu-%s-%s", osName, arch)
}

func FindAssetURL(release *GitHubRelease) (string, error) {
	binaryName := GetBinaryName()

	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no binary found for %s", binaryName)
}
