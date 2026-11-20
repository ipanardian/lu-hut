// Package main initializes the lu-hut CLI application
//
// `lu-hut`is a powerful modern alternative to the Unix ls command that delivers directory listings
// with beautiful box-drawn tables or stunning tree format, intelligent colors, multiple sorting strategies,
// advanced filtering, and seamless git integration. Transform your file exploration from mundane to magnificent.
//
// Coordinate first, complain later.
//
// Copyright (C) 2026
// GitHub: https://github.com/ipanardian/lu-hut
// Author: Ipan Ardian
package main

import (
	"fmt"
	"os"

	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/constants"
	"github.com/ipanardian/lu-hut/internal/lister"
	"github.com/ipanardian/lu-hut/internal/terminal"
	"github.com/ipanardian/lu-hut/internal/updater"
	"github.com/spf13/cobra"
)

func main() {
	go updater.CheckAndNotify()

	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cfg := config.NewDefaultConfig()

	rootCmd := &cobra.Command{
		Use:           "lu [paths...]",
		Short:         "A modern alternative to the Unix ls command with table formatting",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: terminal.Banner(
			constants.Version,
			"A modern alternative to the Unix ls command with box-drawn tables, tree view, intelligent colors, sorting, filtering, and git integration.\n\n"+
				"GitHub: https://github.com/ipanardian/lu-hut",
		),
		Args:    cobra.ArbitraryArgs,
		Version: constants.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				args = []string{"."}
			}

			if err := cfg.Validate(); err != nil {
				return err
			}

			lst := lister.New(cfg)
			return lst.ListPaths(args)
		},
	}

	rootCmd.SetUsageTemplate(terminal.UsageTemplate())

	rootCmd.Flags().StringVar(&cfg.ColorMode, "color", "", "color output mode (always|auto|never)")
	rootCmd.Flags().StringVar(&cfg.IconMode, "icons", "auto", "when to display Nerd Font icons (always|auto|never)")
	rootCmd.Flags().BoolVarP(&cfg.SortModified, "sort-modified", "t", false, "sort by modified time (newest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortSize, "sort-size", "S", false, "sort by file size (largest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortExtension, "sort-extension", "X", false, "sort by file extension")
	rootCmd.Flags().BoolVarP(&cfg.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&cfg.ShowGit, "git", "g", false, "show git status inline")
	rootCmd.Flags().BoolVarP(&cfg.ShowHidden, "hidden", "h", false, "show hidden files")
	rootCmd.Flags().BoolVarP(&cfg.ShowUser, "user", "u", false, "show user and group ownership metadata")
	rootCmd.Flags().BoolVar(&cfg.ShowExactTime, "exact-time", false, "show exact modification time instead of relative")
	rootCmd.Flags().BoolVarP(&cfg.ShowOctal, "octal", "o", false, "show octal permissions instead of rwx")
	rootCmd.Flags().BoolVarP(&cfg.ShowLong, "long", "l", false, "show detailed metadata in tree view (permissions, size, time, user/group)")
	rootCmd.Flags().BoolVarP(&cfg.Tree, "tree", "T", false, "display directory structure in a tree format")
	rootCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "R", false, "list subdirectories recursively")
	rootCmd.Flags().IntVarP(&cfg.MaxDepth, "max-depth", "L", cfg.MaxDepth, "maximum recursion depth (0 = no limit, default: 30)")
	rootCmd.Flags().StringSliceVarP(&cfg.IncludePatterns, "include", "i", nil, "include files matching glob patterns (quote the pattern)")
	rootCmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "x", nil, "exclude files matching glob patterns (quote the pattern)")
	rootCmd.Flags().BoolVarP(&cfg.GitIgnore, "git-ignore", "G", false, "ignore files listed in .gitignore")

	// Cobra's default help flag uses -h as the short form, which collides
	// with --hidden. Register a stub --help flag (no short) up front so
	// Cobra's InitDefaultHelpFlag sees it already exists and skips.
	rootCmd.Flags().Bool("help", false, "help for lu")

	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newRollbackCommand())
	rootCmd.AddCommand(newCompletionCommand(rootCmd))

	return rootCmd
}
