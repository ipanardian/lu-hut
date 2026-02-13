package renderer

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/model"
	"github.com/ipanardian/lu-hut/pkg/helper"
	"golang.org/x/term"
)

func getTerminalWidth() int {
	if width := os.Getenv("COLUMNS"); width != "" {
		if w, err := strconv.Atoi(width); err == nil && w > 0 {
			return w - 10
		}
	}

	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width - 10
	}

	if cmd := exec.Command("tput", "cols"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if w, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil && w > 0 {
				return w - 10
			}
		}
	}

	return 70
}

func calculateDisplayWidths(data [][]string) []int {
	if len(data) == 0 {
		return nil
	}

	widths := make([]int, len(data[0]))

	for _, row := range data {
		for j, cell := range row {
			displayText := helper.StripANSI(cell)
			width := utf8.RuneCountInString(displayText)
			if width > widths[j] {
				widths[j] = width
			}
		}
	}

	return widths
}

func lookupMin(mins []int, idx int, fallback int) int {
	if idx < len(mins) && mins[idx] > 0 {
		return mins[idx]
	}
	return fallback
}

func runeCount(s string) int {
	return utf8.RuneCountInString(s)
}

func truncateMiddle(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runeCount(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}

	runes := []rune(s)
	head := (max - 1) / 2
	tail := max - 1 - head
	if head < 1 {
		head = 1
		tail = max - 1
	}
	if tail < 1 {
		tail = 1
		head = max - 1
	}

	return string(runes[:head]) + "…" + string(runes[len(runes)-tail:])
}

func truncateTail(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runeCount(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}

	runes := []rune(s)
	return "…" + string(runes[len(runes)-(max-1):])
}

const defaultNameMaxWidth = 50

func truncateSymlinkParts(name, target string, maxWidth int) (string, string) {
	if maxWidth <= 0 {
		return "", ""
	}

	arrowLen := runeCount(" -> ")
	if runeCount(name)+arrowLen+runeCount(target) <= maxWidth {
		return name, target
	}

	if maxWidth <= arrowLen+1 {
		return truncateMiddle(name, maxWidth), ""
	}

	targetBudget := maxWidth - arrowLen - runeCount(name)
	if targetBudget >= 2 {
		return name, truncateTail(target, targetBudget)
	}

	remaining := maxWidth - arrowLen
	if remaining <= 1 {
		return truncateMiddle(name, maxWidth), ""
	}

	minTarget := 2
	targetBudget = min(max(remaining/2, minTarget), remaining-1)

	nameBudget := remaining - targetBudget
	if nameBudget < 1 {
		nameBudget = 1
		targetBudget = remaining - nameBudget
	}

	return truncateMiddle(name, nameBudget), truncateTail(target, targetBudget)
}

func formatName(file model.FileEntry, maxWidth int) string {
	originalName := file.Name
	name := originalName
	if maxWidth <= 0 {
		maxWidth = defaultNameMaxWidth
	}

	if file.Mode&fs.ModeSymlink != 0 {
		if target, err := os.Readlink(file.Path); err == nil {
			truncName, truncTarget := truncateSymlinkParts(name, target, maxWidth)
			if truncTarget == "" {
				return color.New(color.FgMagenta, color.Bold).Sprint(truncName)
			}
			return color.New(color.FgMagenta, color.Bold).Sprint(truncName) + " -> " + color.New(color.FgHiBlack).Sprint(truncTarget)
		}
		return color.New(color.FgMagenta, color.Bold).Sprint(truncateMiddle(name, maxWidth))
	}

	name = truncateMiddle(name, maxWidth)

	if file.IsDir {
		return color.New(color.FgBlue, color.Bold).Sprint(name)
	}

	if file.Mode.Perm()&0111 != 0 {
		return color.New(color.FgRed).Sprint(name)
	}

	if file.IsHidden {
		return color.New(color.FgYellow).Sprint(name)
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	switch ext {
	case ".go", ".rs", ".py", ".js", ".ts", ".jsx", ".tsx":
		return color.New(color.FgGreen).Sprint(name)
	case ".md", ".txt", ".rst":
		return color.New(color.FgYellow).Sprint(name)
	case ".yml", ".yaml", ".json", ".toml", ".ini":
		return color.New(color.FgMagenta).Sprint(name)
	default:
		return color.New(color.FgWhite).Sprint(name)
	}
}

func formatSize(size int64, isDir bool) string {
	if isDir {
		return color.New(color.FgCyan).Sprint("-")
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	result := fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])

	return color.New(color.FgHiWhite).Sprint(result)
}

func formatModified(t time.Time, now time.Time, showExact bool) string {
	if showExact {
		c := color.New(color.FgHiWhite)
		return c.Sprint(t.Format("Jan 2, 06 15:04"))
	}

	duration := now.Sub(t)

	var c *color.Color
	var text string

	if duration < 0 {
		c = color.New(color.FgBlue)
		text = "future"
	} else if duration < time.Minute {
		c = color.New(color.FgGreen)
		text = fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		c = color.New(color.FgGreen)
		text = fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		c = color.New(color.FgYellow)
		text = fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else if duration < 7*24*time.Hour {
		c = color.New(color.FgHiYellow)
		text = fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	} else if duration < 30*24*time.Hour {
		c = color.New(color.FgRed)
		text = fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	} else if duration < 365*24*time.Hour {
		c = color.New(color.FgHiRed)
		text = fmt.Sprintf("%d months ago", int(duration.Hours()/(24*30)))
	} else {
		c = color.New(color.FgHiBlack)
		text = fmt.Sprintf("%d years ago", int(duration.Hours()/(24*365)))
	}

	return c.Sprint(text)
}

func formatPermissions(mode fs.FileMode, useOctal bool) string {
	perm := mode.Perm()

	if useOctal {
		return color.New(color.FgHiWhite).Sprint(fmt.Sprintf("%04o", perm))
	}

	var result strings.Builder

	switch {
	case mode&fs.ModeDir != 0:
		result.WriteString(color.New(color.FgCyan, color.Bold).Sprint("d"))
	case mode&fs.ModeSymlink != 0:
		result.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("l"))
	case mode&fs.ModeDevice != 0:
		if mode&fs.ModeCharDevice != 0 {
			result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("c"))
		} else {
			result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("b"))
		}
	case mode&fs.ModeNamedPipe != 0:
		result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("p"))
	case mode&fs.ModeSocket != 0:
		result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("s"))
	default:
		result.WriteString(color.New(color.FgCyan).Sprint("-"))
	}

	for i := 8; i >= 0; i-- {
		bit := perm >> uint(i) & 1
		group := (8 - i) / 3
		var c *color.Color

		switch (8 - i) % 3 {
		case 0:
			if bit == 1 {
				c = color.New(color.FgGreen, color.Bold)
				result.WriteString(c.Sprint("r"))
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		case 1:
			if bit == 1 {
				c = color.New(color.FgYellow, color.Bold)
				result.WriteString(c.Sprint("w"))
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		case 2:
			hasSpecial := false
			switch group {
			case 0:
				hasSpecial = mode&fs.ModeSetuid != 0
			case 1:
				hasSpecial = mode&fs.ModeSetgid != 0
			case 2:
				hasSpecial = mode&fs.ModeSticky != 0
			}

			if hasSpecial {
				if group == 2 {
					if bit == 1 {
						c = color.New(color.FgRed, color.Bold)
						result.WriteString(c.Sprint("t"))
					} else {
						c = color.New(color.FgRed, color.Bold)
						result.WriteString(c.Sprint("T"))
					}
				} else {
					if bit == 1 {
						c = color.New(color.FgMagenta, color.Bold)
						result.WriteString(c.Sprint("s"))
					} else {
						c = color.New(color.FgMagenta, color.Bold)
						result.WriteString(c.Sprint("S"))
					}
				}
			} else if bit == 1 {
				c = color.New(color.FgRed, color.Bold)
				result.WriteString(c.Sprint("x"))
			} else {
				c = color.New(color.FgHiBlack)
				result.WriteString(c.Sprint("-"))
			}
		}
	}

	return result.String()
}

func formatGitStatus(status string) string {
	if status == "" {
		return ""
	}

	switch status {
	case "?":
		return color.New(color.FgRed, color.Bold).Sprint(status)
	case "A", "AM":
		return color.New(color.FgGreen, color.Bold).Sprint(status)
	case "M", " M", "MM":
		return color.New(color.FgYellow, color.Bold).Sprint(status)
	case "D", " D":
		return color.New(color.FgRed).Sprint(status)
	case "R", "C":
		return color.New(color.FgCyan, color.Bold).Sprint(status)
	default:
		return color.New(color.FgYellow).Sprint(status)
	}
}
