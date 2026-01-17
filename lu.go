// Package main provides a modern alternative for the Unix ls command.
// Displays file listings in a beautiful table format with colors,
// git integration, and human-readable file sizes.
//
// Coordinate first, complain later.
//
// GitHub: https://github.com/ipanardian/lu-hutg
// Author: Ipan Ardian
// Version: v1.0.0
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ipanardian/lu-hutg/box/table"
)

type FileEntry struct {
	Name      string
	Path      string
	Size      int64
	Mode      fs.FileMode
	ModTime   time.Time
	IsDir     bool
	IsHidden  bool
	GitStatus string
	Author    string
	Group     string
}

type SortStrategy interface {
	Sort(files []FileEntry, reverse bool)
}

type NameSortStrategy struct{}

func (s *NameSortStrategy) Sort(files []FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		result := strings.Compare(strings.ToLower(files[i].Name), strings.ToLower(files[j].Name))
		if reverse {
			return result > 0
		}
		return result < 0
	})
}

type TimeSortStrategy struct{}

func (s *TimeSortStrategy) Sort(files []FileEntry, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		if reverse {
			return files[i].ModTime.Before(files[j].ModTime)
		}
		return files[i].ModTime.After(files[j].ModTime)
	})
}

type GitRepository struct {
	repoRoot string
	repo     *git.Repository
}

func NewGitRepository(path string) (*GitRepository, error) {
	root, err := findGitRoot(path)
	if err != nil {
		return nil, err
	}
	repo, err := git.PlainOpen(root)
	if err != nil {
		return nil, err
	}
	return &GitRepository{repoRoot: root, repo: repo}, nil
}

func (g *GitRepository) GetStatus(filePath string) string {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return ""
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}

	relPath, err := filepath.Rel(worktree.Filesystem.Root(), absPath)
	if err != nil {
		return ""
	}

	status, err := worktree.Status()
	if err != nil {
		return ""
	}

	fileStatus := status.File(relPath)

	if fileStatus.Worktree == git.Untracked {
		return "??"
	}

	if fileStatus.Worktree == git.Unmodified && fileStatus.Staging == git.Unmodified {
		return ""
	}

	var statusStr string
	if fileStatus.Staging != git.Unmodified {
		switch fileStatus.Staging {
		case git.Added:
			statusStr += "A"
		case git.Modified:
			statusStr += "M"
		case git.Deleted:
			statusStr += "D"
		case git.Renamed:
			statusStr += "R"
		case git.Copied:
			statusStr += "C"
		}
	} else {
		statusStr += " "
	}

	if fileStatus.Worktree != git.Unmodified {
		switch fileStatus.Worktree {
		case git.Modified:
			statusStr += "M"
		case git.Deleted:
			statusStr += "D"
		case git.Added:
			statusStr += "A"
		}
	}

	if statusStr == " " || statusStr == "" {
		return ""
	}

	return strings.TrimSpace(statusStr)
}

type FileFilter struct {
	includePatterns []string
	excludePatterns []string
}

func (f *FileFilter) Apply(files []FileEntry, showHidden bool) []FileEntry {
	var filtered []FileEntry
	for _, file := range files {
		if !showHidden && file.IsHidden {
			continue
		}
		if f.shouldExclude(file.Name) {
			continue
		}
		if len(f.includePatterns) > 0 && !f.shouldInclude(file.Name) {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func (f *FileFilter) shouldExclude(name string) bool {
	for _, pattern := range f.excludePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func (f *FileFilter) shouldInclude(name string) bool {
	for _, pattern := range f.includePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func findGitRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}

type Config struct {
	SortModified    bool
	Reverse         bool
	ShowGit         bool
	ShowHidden      bool
	ShowUser        bool
	IncludePatterns []string
	ExcludePatterns []string
}

type DirectoryLister struct {
	config    Config
	gitRepo   *GitRepository
	filter    *FileFilter
	sortStrat SortStrategy
}

func NewDirectoryLister(config Config) *DirectoryLister {
	filter := &FileFilter{
		includePatterns: config.IncludePatterns,
		excludePatterns: config.ExcludePatterns,
	}

	var sortStrat SortStrategy
	if config.SortModified {
		sortStrat = &TimeSortStrategy{}
	} else {
		sortStrat = &NameSortStrategy{}
	}

	return &DirectoryLister{
		config:    config,
		filter:    filter,
		sortStrat: sortStrat,
	}
}

func (d *DirectoryLister) List(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", absPath)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	if d.config.ShowGit {
		d.gitRepo, _ = NewGitRepository(absPath)
	}

	files := d.collectFiles(absPath, entries)
	files = d.filter.Apply(files, d.config.ShowHidden)
	d.sortStrat.Sort(files, d.config.Reverse)

	renderer := NewTableRenderer(d.config)
	renderer.Render(files, time.Now())

	return nil
}

func (d *DirectoryLister) collectFiles(path string, entries []fs.DirEntry) []FileEntry {
	files := make([]FileEntry, 0, len(entries))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := FileEntry{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Size:     info.Size(),
			Mode:     info.Mode(),
			ModTime:  info.ModTime(),
			IsDir:    entry.IsDir(),
			IsHidden: strings.HasPrefix(entry.Name(), "."),
		}

		if d.config.ShowGit && d.gitRepo != nil {
			file.GitStatus = d.gitRepo.GetStatus(file.Path)
		}

		if d.config.ShowUser {
			file.Author, file.Group = extractUserGroup(info)
		}

		files = append(files, file)
	}

	return files
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	var config Config

	rootCmd := &cobra.Command{
		Use:   "lu [path]",
		Short: "A modern alternative to the Unix ls command with table formatting",
		Long: `lu-hutg is a modern alternative to the Unix ls command with box-drawn tables, colors, and git integration.

GitHub: https://github.com/ipanardian/lu-hutg
Version: v1.0.0`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			if path != "." {
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					if len(config.IncludePatterns) > 0 {
						config.IncludePatterns = append(config.IncludePatterns, path)
						path = "."
					}
				}
			}

			lister := NewDirectoryLister(config)
			return lister.List(path)
		},
	}

	rootCmd.Flags().BoolP("help", "", false, "help for lu")
	rootCmd.Flags().BoolVarP(&config.SortModified, "sort-modified", "t", false, "sort by modified time (newest first)")
	rootCmd.Flags().BoolVarP(&config.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&config.ShowGit, "git", "g", false, "show git status inline")
	rootCmd.Flags().BoolVarP(&config.ShowHidden, "hidden", "h", false, "show hidden files")
	rootCmd.Flags().BoolVarP(&config.ShowUser, "user", "u", false, "show user & group")
	rootCmd.Flags().StringSliceVarP(&config.IncludePatterns, "include", "i", nil, "include files matching glob patterns (quote the pattern)")
	rootCmd.Flags().StringSliceVarP(&config.ExcludePatterns, "exclude", "x", nil, "exclude files matching glob patterns (quote the pattern)")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		showColoredHelp(cmd)
	})

	return rootCmd
}

type TableRenderer struct {
	config Config
}

func NewTableRenderer(config Config) *TableRenderer {
	return &TableRenderer{config: config}
}

func (r *TableRenderer) Render(files []FileEntry, now time.Time) {
	if len(files) == 0 {
		return
	}

	terminalWidth := max(getTerminalWidth(), 40)

	data := r.buildTableData(files, now)
	displayWidths := calculateDisplayWidths(data)
	mins, maxs := r.columnConstraints()

	for i := range displayWidths {
		if i < len(mins) && mins[i] > 0 && displayWidths[i] < mins[i] {
			displayWidths[i] = mins[i]
		}
		if i < len(maxs) && maxs[i] > 0 && displayWidths[i] > maxs[i] {
			displayWidths[i] = maxs[i]
		}
	}

	minContentWidth := 0
	for i := range displayWidths {
		minContentWidth += lookupMin(mins, i, 4)
	}
	minBorderWidth := (len(displayWidths)-1)*3 + 2
	if terminalWidth < minContentWidth+minBorderWidth {
		fmt.Println("Terminal is too small to display the table. Please widen your terminal window.")
		return
	}

	totalContentWidth := 0
	for _, w := range displayWidths {
		totalContentWidth += w
	}
	borderWidth := (len(displayWidths)-1)*3 + 2
	totalWidth := totalContentWidth + borderWidth

	if totalWidth > terminalWidth {
		r.shrinkColumns(displayWidths, mins, totalWidth-terminalWidth)
	}

	tbl := table.NewTableWithWidths(data, displayWidths)
	tbl.SetBorderStyle(0)
	tbl.SetHeaderStyle(1)
	tbl.SetHeaderColor(color.New(color.FgWhite, color.Bold))
	tbl.SetBorderColor(color.New(color.FgGreen))
	tbl.Print()
}

func (r *TableRenderer) buildTableData(files []FileEntry, now time.Time) [][]string {
	headers := []string{"Name", "Size", "Modified", "Perms"}
	if r.config.ShowGit {
		headers = append(headers, "Git")
	}
	if r.config.ShowUser {
		headers = append(headers, "User", "Group")
	}

	data := make([][]string, len(files)+1)
	data[0] = headers

	for i, file := range files {
		row := []string{
			formatName(file),
			formatSize(file.Size, file.IsDir),
			formatModified(file.ModTime, now),
			formatPermissions(file.Mode),
		}
		if r.config.ShowGit {
			row = append(row, formatGitStatus(file.GitStatus))
		}
		if r.config.ShowUser {
			row = append(row, file.Author, file.Group)
		}
		data[i+1] = row
	}

	return data
}

func (r *TableRenderer) columnConstraints() ([]int, []int) {
	mins := []int{15, 6, 10, 10}
	maxs := []int{50, 10, 15, 12}
	if r.config.ShowGit {
		mins = append(mins, 6)
		maxs = append(maxs, 12)
	}
	if r.config.ShowUser {
		mins = append(mins, 6, 6)
		maxs = append(maxs, 12, 12)
	}
	return mins, maxs
}

func (r *TableRenderer) shrinkColumns(displayWidths, mins []int, excess int) {
	totalShrinkable := 0
	for i, w := range displayWidths {
		if i != 1 && i != 3 {
			minWidth := lookupMin(mins, i, 4)
			if w-minWidth > 0 {
				totalShrinkable += w - minWidth
			}
		}
	}

	for i := range displayWidths {
		if i != 1 && i != 3 {
			minWidth := lookupMin(mins, i, 4)
			shrinkable := displayWidths[i] - minWidth
			if shrinkable > 0 && totalShrinkable > 0 {
				shrinkAmount := (shrinkable * excess) / totalShrinkable
				shrinkAmount = min(shrinkAmount, shrinkable)
				displayWidths[i] -= shrinkAmount
				displayWidths[i] = max(displayWidths[i], minWidth)
				excess -= shrinkAmount
				totalShrinkable -= shrinkable
			}
		}
	}
}

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

func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' {
			j := i + 1
			if j < len(s) && s[j] == '[' {
				j++
				for j < len(s) && (s[j] < 'a' || s[j] > 'z') && (s[j] < 'A' || s[j] > 'Z') {
					j++
				}
				j++
			}
			i = j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

func calculateDisplayWidths(data [][]string) []int {
	if len(data) == 0 {
		return nil
	}

	widths := make([]int, len(data[0]))

	for _, row := range data {
		for j, cell := range row {
			displayText := stripANSI(cell)
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

func formatName(file FileEntry) string {
	name := file.Name

	if file.IsDir {
		return color.New(color.FgBlue, color.Bold).Sprint(name)
	}

	if file.Mode.Perm()&0111 != 0 {
		return color.New(color.FgRed).Sprint(name)
	}

	if file.IsHidden {
		return color.New(color.FgYellow).Sprint(name)
	}

	ext := strings.ToLower(filepath.Ext(name))
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

	units := []string{"KB", "MB", "GB", "TB"}
	result := fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])

	return color.New(color.FgHiWhite).Sprint(result)
}

func formatModified(t time.Time, now time.Time) string {
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

func formatPermissions(mode fs.FileMode) string {
	perm := mode.Perm()

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
			if bit == 1 {
				if mode&fs.ModeSetuid != 0 {
					c = color.New(color.FgMagenta, color.Bold)
					result.WriteString(c.Sprint("s"))
				} else if mode&fs.ModeSetgid != 0 {
					c = color.New(color.FgMagenta, color.Bold)
					result.WriteString(c.Sprint("s"))
				} else if mode&fs.ModeSticky != 0 {
					c = color.New(color.FgRed, color.Bold)
					result.WriteString(c.Sprint("t"))
				} else {
					c = color.New(color.FgRed, color.Bold)
					result.WriteString(c.Sprint("x"))
				}
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
	case "??":
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

func extractUserGroup(fileInfo os.FileInfo) (string, string) {
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		u, errU := user.LookupId(strconv.Itoa(int(stat.Uid)))
		g, errG := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

		username := "unknown"
		groupname := "unknown"

		if errU == nil {
			username = u.Username
		}
		if errG == nil {
			groupname = g.Name
		}

		return color.New(color.FgWhite).Sprint(username), color.New(color.FgWhite).Sprint(groupname)
	}
	return color.New(color.FgWhite).Sprint("unknown"), color.New(color.FgWhite).Sprint("unknown")
}

func showColoredHelp(_ *cobra.Command) {
	fmt.Printf("\n%s %s\n\n",
		color.New(color.FgCyan, color.Bold).Sprint("lu-hutg v1.0.0"),
		color.New(color.FgHiWhite).Sprint("- a modern alternative to the Unix ls command with box-drawn tables, colors, and git integration"),
	)
	fmt.Printf("%s\n\n", color.New(color.FgHiBlack).Sprint("GitHub: https://github.com/ipanardian/lu-hutg"))

	fmt.Printf("%s\n\n", color.New(color.FgWhite).Sprint("USAGE:"))
	fmt.Printf("  lu [path] [flags]\n\n")

	fmt.Printf("%s\n", color.New(color.FgWhite, color.Bold).Sprint("FLAGS:"))

	flags := []struct {
		flag, desc string
	}{
		{"-t, --sort-modified", "sort by modified time (newest first)"},
		{"-r, --reverse", "reverse sort order"},
		{"-g, --git", "show git status inline"},
		{"-a, --author", "show file author/owner"},
		{"-h, --hidden", "show hidden files"},
		{"-u, --unit", "show unit coordination (user & group)"},
		{"-i, --include", "include files matching glob patterns (quote the pattern)"},
		{"-x, --exclude", "exclude files matching glob patterns (quote the pattern)"},
		{"--help", "show this help message"},
	}

	for _, f := range flags {
		fmt.Printf("  %s\t%s\n",
			color.New(color.FgCyan, color.Bold).Sprintf("%-20s", f.flag),
			color.New(color.FgHiWhite).Sprint(f.desc),
		)
	}

	fmt.Printf("\n%s\n", color.New(color.FgWhite, color.Bold).Sprint("EXAMPLES:"))
	examples := []string{
		"lu",
		"lu -t",
		"lu -tr",
		"lu -g",
		"lu -tg",
		"lu -a",
		"lu -ta",
		"lu -i '*.go'",
		"lu -x '*.tmp'",
		"lu -hutg",
	}

	for _, ex := range examples {
		fmt.Printf("  %s\n", color.New(color.FgGreen).Sprint(ex))
	}

	fmt.Println()
}
