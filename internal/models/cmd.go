package models

import (
	"fmt"
	"strings"
)

// AppStatus describes why an installed app needs attention.
type AppStatus string

const (
	StatusUnknown     AppStatus = ""
	StatusOutdated    AppStatus = "outdated"
	StatusHeld        AppStatus = "held"
	StatusFailed      AppStatus = "failed"
	StatusDeprecated  AppStatus = "deprecated"
	StatusRemoved     AppStatus = "removed"
	StatusMissingDeps AppStatus = "missing-deps"
)

// Label returns the user-facing Chinese label for a status.
func (s AppStatus) Label() string {
	switch s {
	case StatusOutdated:
		return "可更新"
	case StatusHeld:
		return "已锁定"
	case StatusFailed:
		return "安装失败"
	case StatusDeprecated:
		return "已废弃"
	case StatusRemoved:
		return "清单已移除"
	case StatusMissingDeps:
		return "缺少依赖"
	default:
		return ""
	}
}

// AppInfo is the unified representation of a package across sfsu and scoop.
type AppInfo struct {
	Name             string    `json:"name"`
	Version          string    `json:"version,omitempty"`
	CurrentVersion   string    `json:"current_version,omitempty"`
	AvailableVersion string    `json:"available_version,omitempty"`
	Bucket           string    `json:"bucket,omitempty"`
	Source           string    `json:"source,omitempty"`
	Installed        bool      `json:"installed"`
	Global           bool      `json:"global,omitempty"`
	Status           AppStatus `json:"status,omitempty"`
	Info             string    `json:"info,omitempty"`
	Notes            string    `json:"notes,omitempty"`
}

// Origin returns the bucket or source name without the app name.
func (a AppInfo) Origin() string {
	if a.Bucket != "" {
		return a.Bucket
	}
	return a.Source
}

// FullName returns the bucket/name format accepted by scoop install.
func (a AppInfo) FullName() string {
	if b := a.Origin(); b != "" {
		return fmt.Sprintf("%s/%s", b, a.Name)
	}
	return a.Name
}

// ID implements tui.Item. Global installs get a distinct identity so the
// same app can be selected locally and globally at the same time.
func (a AppInfo) ID() string {
	id := a.FullName()
	if id == "" {
		id = a.Name
	}
	if a.Installed && a.Global {
		id += "@global"
	}
	return id
}

// Title implements tui.Item and the bubbletea list.Item interface.
func (a AppInfo) Title() string {
	return a.Name
}

// FilterValue implements the bubbletea list.Item interface.
func (a AppInfo) FilterValue() string {
	return a.Name
}

// Description implements tui.Item and the bubbletea list.Item interface.
func (a AppInfo) Description() string {
	parts := make([]string, 0, 4)
	if origin := a.Origin(); origin != "" {
		parts = append(parts, origin)
	}
	if v := a.DisplayVersion(); v != "" {
		parts = append(parts, v)
	}
	if a.Installed {
		parts = append(parts, "已安装")
	}
	if a.Global {
		parts = append(parts, "全局")
	}
	if label := a.Status.Label(); label != "" {
		parts = append(parts, label)
	}
	if a.Info != "" && a.Status == StatusUnknown {
		parts = append(parts, oneLine(a.Info))
	}
	if a.Notes != "" && a.Status == StatusUnknown && a.Info == "" {
		parts = append(parts, oneLine(a.Notes))
	}
	return strings.Join(parts, " · ")
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.Join(strings.Fields(s), " ")
}

// DisplayVersion renders current and available versions without losing either.
func (a AppInfo) DisplayVersion() string {
	if a.CurrentVersion != "" && a.AvailableVersion != "" && a.CurrentVersion != a.AvailableVersion {
		return fmt.Sprintf("%s → %s", a.CurrentVersion, a.AvailableVersion)
	}
	if a.AvailableVersion != "" {
		return a.AvailableVersion
	}
	if a.CurrentVersion != "" {
		return a.CurrentVersion
	}
	return a.Version
}

// HasUpdate reports whether the app has a newer version available.
func (a AppInfo) HasUpdate() bool {
	return a.Status == StatusOutdated ||
		(a.AvailableVersion != "" && a.AvailableVersion != a.CurrentVersion)
}
