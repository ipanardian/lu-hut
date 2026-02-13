// Package config provides configuration management for the lu-hut application.
package config

import "fmt"

type Config struct {
	SortModified    bool
	SortSize        bool
	SortExtension   bool
	Reverse         bool
	ShowGit         bool
	ShowHidden      bool
	ShowUser        bool
	ShowExactTime   bool
	ShowOctal       bool
	Recursive       bool
	Tree            bool
	MaxDepth        int
	ColorMode       string
	IncludePatterns []string
	ExcludePatterns []string
}

func NewDefaultConfig() Config {
	return Config{
		MaxDepth: 30,
	}
}

func (c Config) Validate() error {
	if c.MaxDepth < 0 {
		return fmt.Errorf("max depth cannot be negative")
	}
	if c.ColorMode != "" && c.ColorMode != "always" && c.ColorMode != "auto" && c.ColorMode != "never" {
		return fmt.Errorf("invalid color mode: %s (must be always, auto, or never)", c.ColorMode)
	}
	return nil
}
