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
	ShowLong        bool
	Recursive       bool
	Tree            bool
	GitIgnore       bool
	MaxDepth        int
	ColorMode       string
	IconMode        string
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
	if c.IconMode != "" && c.IconMode != "always" && c.IconMode != "auto" && c.IconMode != "never" {
		return fmt.Errorf("invalid icon mode: %s (must be always, auto, or never)", c.IconMode)
	}
	return nil
}
