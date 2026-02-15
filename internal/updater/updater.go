package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

const downloadTimeout = 5 * time.Minute

func PerformUpdate(release *GitHubRelease) error {
	downloadURL, err := FindAssetURL(release)
	if err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "lu-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	color.Cyan("Downloading %s...", release.TagName)

	client := &http.Client{
		Timeout: downloadTimeout,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmpFile.Close()

	if written == 0 {
		return fmt.Errorf("downloaded file is empty")
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		if err := os.Rename(backupPath, execPath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	color.Green("✓ Successfully updated to %s", release.TagName)
	color.Yellow("→ Previous version backed up (use 'lu rollback' to restore)")
	return nil
}

func PerformRollback() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	backupPath := execPath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	tmpPath := execPath + ".tmp"
	if err := os.Rename(execPath, tmpPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := os.Rename(backupPath, execPath); err != nil {
		if restoreErr := os.Rename(tmpPath, execPath); restoreErr != nil {
			return fmt.Errorf("failed to restore backup and rollback failed: %w", err)
		}
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	_ = os.Remove(tmpPath)

	color.Green("✓ Successfully rolled back to previous version")
	return nil
}

func CheckAndNotify() {
	cacheFile := getCacheFilePath()

	if shouldSkipCheck(cacheFile) {
		return
	}

	release, err := GetLatestVersion()
	if err != nil {
		return
	}

	updateCacheFile(cacheFile)

	currentVersion := GetCurrentVersion()
	if IsNewerVersion(currentVersion, release.TagName) {
		yellow := color.New(color.FgYellow).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Fprintf(os.Stderr, "\n%s New version %s available (current: %s)\n",
			yellow("⚠"),
			cyan(release.TagName),
			currentVersion)
		fmt.Fprintf(os.Stderr, "%s Run %s to upgrade\n\n",
			yellow("→"),
			cyan("lu update"))
	}
}

func getCacheFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	cacheDir := filepath.Join(homeDir, ".lu-hut")
	_ = os.MkdirAll(cacheDir, 0755)

	return filepath.Join(cacheDir, "last_check")
}

func shouldSkipCheck(cacheFile string) bool {
	if cacheFile == "" {
		return true
	}

	info, err := os.Stat(cacheFile)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < 24*time.Hour
}

func updateCacheFile(cacheFile string) {
	if cacheFile == "" {
		return
	}

	_ = os.WriteFile(cacheFile, []byte(time.Now().Format(time.RFC3339)), 0644)
}
