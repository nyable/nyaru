package models

import (
	"fmt"
)

// CmdAction struct
type CmdAction struct {
	Command string
	Desc    string
}

// AppInfo represents unified information about a scoop package
type AppInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Bucket    string `json:"bucket"`
	Source    string `json:"source"`
	Installed bool   `json:"installed"`
	Selected  bool   // Internal field for multiselect
}

// FullName returns the Bucket/Name format
func (a *AppInfo) FullName() string {
	b := a.Bucket
	if b == "" {
		b = a.Source
	}
	if b != "" {
		return fmt.Sprintf("%s/%s", b, a.Name)
	}
	return a.Name
}

// FilterValue implements the bubbletea list.Item interface
func (a AppInfo) FilterValue() string {
	return a.Name
}

// Title implements the bubbletea list.DefaultItem interface
func (a AppInfo) Title() string {
	return a.Name
}

// Description implements the bubbletea list.DefaultItem interface
func (a AppInfo) Description() string {
	status := ""
	if a.Installed {
		status = " [Installed]"
	}
	selected := ""
	if a.Selected {
		selected = " [√]"
	}
	return fmt.Sprintf("%s v%s%s%s", a.FullName(), a.Version, status, selected)
}
