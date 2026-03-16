package models

import (
	"fmt"
)

type CacheResult struct {
	Index      int    `json:"-"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Length     int64  `json:"length"`
	FormatSize string `json:"-"`
	Selected   bool   `json:"-"`
}

// FilterValue implements the bubbletea list.Item interface
func (c CacheResult) FilterValue() string {
	return c.Name
}

// Title implements the bubbletea list.DefaultItem interface
func (c CacheResult) Title() string {
	return c.Name
}

// Description implements the bubbletea list.DefaultItem interface
func (c CacheResult) Description() string {
	if c.FormatSize == "" {
		// This should be pre-calculated by the parser, but for safety:
		// (wait, I removed the utils dependency, but the parser already sets it)
	}
	selected := ""
	if c.Selected {
		selected = " [√]"
	}
	return fmt.Sprintf("v%s - %s%s", c.Version, c.FormatSize, selected)
}
