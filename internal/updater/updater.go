package updater

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

const downloadTimeout = 5 * time.Minute

// binaryName returns the binary filename inside release archives for the
// current platform. GoReleaser names Windows binaries with a .exe suffix.
func binaryName() string {
	if runtime.GOOS == "windows" {
		return "lu.exe"
	}
	return "lu"
}

// renameWithReplace moves src to dst, replacing dst if it already exists.
// os.Rename on Windows fails when dst exists (unlike the atomic replace on
// Unix), so an existing destination must be removed first.
func renameWithReplace(src, dst string) error {
	if runtime.GOOS == "windows" {
		if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return os.Rename(src, dst)
}

func IsHomebrewInstallation() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	execPath, _ = filepath.EvalSymlinks(execPath)

	return strings.Contains(execPath, "/Cellar/") ||
		strings.Contains(execPath, "homebrew") ||
		strings.Contains(execPath, "/opt/homebrew/") ||
		strings.Contains(execPath, "/usr/local/Cellar/")
}

func PerformUpdate(release *GitHubRelease) error {
	if IsHomebrewInstallation() {
		color.Yellow("⚠ lu-hut was installed via Homebrew")
		color.Cyan("→ Please use 'brew upgrade lu-hut' to update")
		return nil
	}

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

	color.Cyan("Downloading %s...", release.TagName)

	client := &http.Client{
		Timeout: downloadTimeout,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	color.Cyan("Extracting binary...")

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var binaryData []byte
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == binaryName() {
			binaryData, err = io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("failed to read binary from archive: %w", err)
			}
			break
		}
	}

	if len(binaryData) == 0 {
		return fmt.Errorf("binary %q not found in archive", binaryName())
	}

	tmpFile, err := os.CreateTemp("", "lu-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(binaryData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	backupPath := execPath + ".backup"
	if err := renameWithReplace(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := renameWithReplace(tmpPath, execPath); err != nil {
		if restoreErr := renameWithReplace(backupPath, execPath); restoreErr != nil {
			return fmt.Errorf("failed to replace binary and failed to restore backup (restore: %v): %w", restoreErr, err)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	color.Green("✓ Successfully updated to %s", release.TagName)
	color.Yellow("→ Previous version backed up (use 'lu rollback' to restore)")
	return nil
}

func PerformRollback() error {
	if IsHomebrewInstallation() {
		color.Yellow("⚠ lu-hut was installed via Homebrew")
		color.Cyan("→ Rollback is not supported for Homebrew installations")
		color.Cyan("→ Use 'brew install lu-hut@<version>' to install a specific version")
		return fmt.Errorf("rollback is not supported for Homebrew installations")
	}

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
	if err := renameWithReplace(execPath, tmpPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := renameWithReplace(backupPath, execPath); err != nil {
		if restoreErr := renameWithReplace(tmpPath, execPath); restoreErr != nil {
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

		if IsHomebrewInstallation() {
			fmt.Fprintf(os.Stderr, "%s Run %s to upgrade\n\n",
				yellow("→"),
				cyan("brew upgrade lu-hut"))
		} else {
			fmt.Fprintf(os.Stderr, "%s Run %s to upgrade\n\n",
				yellow("→"),
				cyan("lu update"))
		}
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
