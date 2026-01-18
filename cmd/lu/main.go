// Package main provides a modern alternative for the Unix ls command.
// Displays file listings in a beautiful table format with colors,
// filtering, git integration, and human-readable file sizes.
//
// Coordinate first, complain later.
//
// GitHub: https://github.com/ipanardian/lu-hut
// Author: Ipan Ardian
// Version: v1.0.0
package main

import (
	"log"
	"os"

	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/lister"
	"github.com/ipanardian/lu-hut/internal/terminal"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	var cfg config.Config

	rootCmd := &cobra.Command{
		Use:   "lu [path]",
		Short: "A modern alternative to the Unix ls command with table formatting",
		Long: `lu-hut is a modern alternative to the Unix ls command with box-drawn tables, colors, filtering, and git integration.

GitHub: https://github.com/ipanardian/lu-hut
Version: v1.0.0`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			if path != "." {
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					if len(cfg.IncludePatterns) > 0 {
						cfg.IncludePatterns = append(cfg.IncludePatterns, path)
						path = "."
					}
				}
			}

			lister := lister.New(cfg)
			return lister.List(path)
		},
	}

	rootCmd.Flags().BoolP("help", "", false, "help for lu")
	rootCmd.Flags().BoolVarP(&cfg.SortModified, "sort-modified", "t", false, "sort by modified time (newest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortSize, "sort-size", "S", false, "sort by file size (largest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortExtension, "sort-extension", "X", false, "sort by file extension")
	rootCmd.Flags().BoolVarP(&cfg.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&cfg.ShowGit, "git", "g", false, "show git status inline")
	rootCmd.Flags().BoolVarP(&cfg.ShowHidden, "hidden", "h", false, "show hidden files")
	rootCmd.Flags().BoolVarP(&cfg.ShowUser, "user", "u", false, "show user and group ownership metadata")
	rootCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "R", false, "list subdirectories recursively")
	rootCmd.Flags().IntVarP(&cfg.MaxDepth, "max-depth", "L", 0, "maximum recursion depth (0 = no limit, default: 30)")
	rootCmd.Flags().StringSliceVarP(&cfg.IncludePatterns, "include", "i", nil, "include files matching glob patterns (quote the pattern)")
	rootCmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "x", nil, "exclude files matching glob patterns (quote the pattern)")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		terminal.ShowColoredHelp(cmd)
	})

	return rootCmd
}
