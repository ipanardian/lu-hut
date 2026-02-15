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
	"log"
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
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	cfg := config.NewDefaultConfig()

	rootCmd := &cobra.Command{
		Use:   "lu [path]",
		Short: "A modern alternative to the Unix ls command with table formatting",
		Long: `lu-hut is a powerful modern alternative to the Unix ls command with beautiful box-drawn tables or stunning tree format, intelligent colors, multiple sorting strategies, advanced filtering, and seamless git integration.

GitHub: https://github.com/ipanardian/lu-hut
Version: ` + constants.Version,
		Args:    cobra.MaximumNArgs(1),
		Version: constants.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			if err := cfg.Validate(); err != nil {
				return err
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

	rootCmd.Flags().StringVar(&cfg.ColorMode, "color", "", "color output mode (always|auto|never)")
	rootCmd.Flags().BoolVarP(&cfg.SortModified, "sort-modified", "t", false, "sort by modified time (newest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortSize, "sort-size", "S", false, "sort by file size (largest first)")
	rootCmd.Flags().BoolVarP(&cfg.SortExtension, "sort-extension", "X", false, "sort by file extension")
	rootCmd.Flags().BoolVarP(&cfg.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&cfg.ShowGit, "git", "g", false, "show git status inline")
	rootCmd.Flags().BoolVarP(&cfg.ShowHidden, "hidden", "h", false, "show hidden files")
	rootCmd.Flags().BoolVarP(&cfg.ShowUser, "user", "u", false, "show user and group ownership metadata")
	rootCmd.Flags().BoolVarP(&cfg.ShowExactTime, "exact-time", "T", false, "show exact modification time instead of relative")
	rootCmd.Flags().BoolVarP(&cfg.ShowOctal, "octal", "o", false, "show octal permissions instead of rwx")
	rootCmd.Flags().BoolVarP(&cfg.Tree, "tree", "F", false, "display directory structure in a tree format")
	rootCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "R", false, "list subdirectories recursively")
	rootCmd.Flags().IntVarP(&cfg.MaxDepth, "max-depth", "L", cfg.MaxDepth, "maximum recursion depth (0 = no limit, default: 30)")
	rootCmd.Flags().StringSliceVarP(&cfg.IncludePatterns, "include", "i", nil, "include files matching glob patterns (quote the pattern)")
	rootCmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "x", nil, "exclude files matching glob patterns (quote the pattern)")

	var help bool
	rootCmd.Flags().BoolVar(&help, "help", false, "help for lu")
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		terminal.ShowColoredHelp(cmd)
	})

	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newRollbackCommand())

	return rootCmd
}
