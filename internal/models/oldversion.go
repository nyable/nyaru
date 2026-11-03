package models

import (
	"fmt"
	"strings"
)

// OldVersionEntry describes one obsolete version directory that scoop cleanup
// can remove. Sizes are filled in by the parser so both list and detail views
// can show how much space a cleanup would free.
type OldVersionEntry struct {
	Version    string `json:"version"`
	Length     int64  `json:"length"`
	FormatSize string `json:"format_size,omitempty"`
}

// OldVersion groups the obsolete versions of one installed application.
// Local and global installations are separate entries because scoop cleanup
// needs the matching --global flag.
type OldVersion struct {
	Name       string            `json:"name"`
	Current    string            `json:"current,omitempty"`
	Global     bool              `json:"global,omitempty"`
	Versions   []OldVersionEntry `json:"versions,omitempty"`
	Length     int64             `json:"length"`
	FormatSize string            `json:"format_size,omitempty"`
}

// ID implements tui.Item and keeps local/global entries distinct.
func (o OldVersion) ID() string {
	id := o.Name
	if o.Global {
		id += "@global"
	}
	return id
}

// Title implements tui.Item.
func (o OldVersion) Title() string { return o.Name }

// FilterValue implements tui.Item.
func (o OldVersion) FilterValue() string { return o.Name }

// Description implements tui.Item and lists exactly what will be reclaimed.
func (o OldVersion) Description() string {
	parts := make([]string, 0, 5)
	if o.Current != "" {
		parts = append(parts, "当前 "+o.Current)
	}
	names := make([]string, 0, len(o.Versions))
	for _, v := range o.Versions {
		names = append(names, v.Version)
	}
	if len(names) > 0 {
		parts = append(parts, fmt.Sprintf("旧版本 %d 个 (%s)", len(names), strings.Join(names, ", ")))
	}
	if o.FormatSize != "" {
		parts = append(parts, "可释放 "+o.FormatSize)
	}
	if o.Global {
		parts = append(parts, "全局")
	}
	return strings.Join(parts, " · ")
}

func (o OldVersion) String() string {
	return fmt.Sprintf("%s (%s)", o.Name, o.Description())
}
