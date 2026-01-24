// Package terminal provides terminal-related utilities like help display.
package terminal

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/constants"
	"github.com/spf13/cobra"
)

func ShowColoredHelp(_ *cobra.Command) {
	fmt.Printf("\n%s %s\n\n",
		color.New(color.FgCyan, color.Bold).Sprint("lu-hut "+constants.Version),
		color.New(color.FgHiWhite).Sprint("- a modern alternative to the Unix ls command with box-drawn tables, tree-view, colors, filtering, sorting and git integration"),
	)
	fmt.Printf("%s\n\n", color.New(color.FgHiBlack).Sprint("GitHub: https://github.com/ipanardian/lu-hut"))

	fmt.Printf("%s\n\n", color.New(color.FgWhite).Sprint("USAGE:"))
	fmt.Printf("  lu [path] [flags]\n\n")

	fmt.Printf("%s\n", color.New(color.FgWhite, color.Bold).Sprint("FLAGS:"))

	flags := []struct {
		flag, desc string
	}{
		{"-t, --sort-modified", "sort by modified time (newest first)"},
		{"-S, --sort-size", "sort by file size (largest first)"},
		{"-X, --sort-extension", "sort by file extension"},
		{"-r, --reverse", "reverse sort order"},
		{"-g, --git", "show git status inline"},
		{"-h, --hidden", "show hidden files"},
		{"-u, --user", "show user and group ownership metadata."},
		{"-T, --exact-time", "show exact modification time instead of relative"},
		{"-F, --tree", "display directory structure in a tree format."},
		{"-R, --recursive", "list subdirectories recursively"},
		{"-L, --max-depth", "maximum recursion depth (0 = no limit, default: 30)"},
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
		"lu -S",
		"lu -F",
		"lu -i '*.go'",
		"lu -x '*.tambang'",
		"lu -hut (Lord's mode)",
	}

	for _, ex := range examples {
		fmt.Printf("  %s\n", color.New(color.FgGreen).Sprint(ex))
	}

	fmt.Println()
}
