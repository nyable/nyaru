package models

import (
	"fmt"
	"strings"
)

// CacheResult is one cache entry grouped by application. Scoop's cache rm
// command operates on an application, not on an individual version, so the
// UI mirrors that granularity.
type CacheResult struct {
	Index      int      `json:"-"`
	App        string   `json:"app"`
	Versions   []string `json:"versions,omitempty"`
	Length     int64    `json:"length"`
	FormatSize string   `json:"format_size,omitempty"`
}

// ID implements tui.Item.
func (c CacheResult) ID() string { return c.App }

// Title implements tui.Item and the bubbletea list.Item interface.
func (c CacheResult) Title() string { return c.App }

// FilterValue implements the bubbletea list.Item interface.
func (c CacheResult) FilterValue() string { return c.App }

// Description implements tui.Item and the bubbletea list.Item interface.
func (c CacheResult) Description() string {
	parts := make([]string, 0, 2)
	if len(c.Versions) > 0 {
		parts = append(parts, fmt.Sprintf("%d 个版本 (%s)", len(c.Versions), strings.Join(c.Versions, ", ")))
	}
	if c.FormatSize != "" {
		parts = append(parts, "大小 "+c.FormatSize)
	}
	return strings.Join(parts, " · ")
}

func (c CacheResult) String() string {
	return fmt.Sprintf("%s (%s)", c.App, c.Description())
}
