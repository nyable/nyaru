package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsValues are the nyaru settings that can be changed inside the TUI.
// The CLI converts them to and from config.Config.
type SettingsValues struct {
	Mode          string
	ConfirmBatch  bool
	StatusTimeout int // seconds; 0 keeps status messages until the next screen
}

const defaultStatusTimeout = 5

func DefaultSettings() SettingsValues {
	return SettingsValues{
		Mode:          "auto",
		ConfirmBatch:  true,
		StatusTimeout: defaultStatusTimeout,
	}
}

func normalizeSettings(s SettingsValues) SettingsValues {
	switch strings.ToLower(strings.TrimSpace(s.Mode)) {
	case "sfsu", "scoop":
		s.Mode = strings.ToLower(strings.TrimSpace(s.Mode))
	default:
		s.Mode = "auto"
	}
	if s.StatusTimeout < 0 {
		s.StatusTimeout = defaultStatusTimeout
	}
	if s.StatusTimeout > 3600 {
		s.StatusTimeout = 3600
	}
	return s
}

func (s SettingsValues) statusTTL() time.Duration {
	if s.StatusTimeout <= 0 {
		return 0
	}
	return time.Duration(s.StatusTimeout) * time.Second
}

type settingsItem struct {
	key   string
	label string
	value string
	desc  string
}

func (a *App) settingsItems() []settingsItem {
	statusValue := fmt.Sprintf("%d 秒", a.settings.StatusTimeout)
	if a.settings.StatusTimeout == 0 {
		statusValue = "不自动清除"
	}
	return []settingsItem{
		{
			key:   "mode",
			label: "后端模式",
			value: a.modeLabel(),
			desc:  "auto 会优先使用已安装的 sfsu；修改后重启 nyaru 生效",
		},
		{
			key:   "confirm_batch",
			label: "批量操作确认",
			value: onOff(a.settings.ConfirmBatch),
			desc:  "批量安装、更新、清理前是否显示确认提示",
		},
		{
			key:   "status_timeout",
			label: "状态提示时长",
			value: statusValue,
			desc:  "底部执行结果提示自动消失的时间；0 表示不自动清除",
		},
		{
			key:   "reset",
			label: "恢复默认设置",
			value: "Enter 执行",
			desc:  "恢复 auto、批量确认开启、状态提示 5 秒",
		},
	}
}

func (a *App) modeLabel() string {
	switch a.settings.Mode {
	case "sfsu":
		return "sfsu"
	case "scoop":
		return "scoop"
	default:
		if a.backend != nil && a.backend.Name() != "" {
			return "自动（当前 " + a.backend.Name() + "）"
		}
		return "自动"
	}
}

func onOff(v bool) string {
	if v {
		return "开启"
	}
	return "关闭"
}

func (a *App) settingsView() string {
	items := a.settingsItems()
	var b strings.Builder
	b.WriteString(a.theme.Subtitle.Render("修改后立即保存；后端模式在下次启动时生效。"))
	b.WriteString("\n\n")

	labelWidth := 0
	for _, item := range items {
		if w := lipgloss.Width(item.label); w > labelWidth {
			labelWidth = w
		}
	}
	for i, item := range items {
		cursor := "  "
		if i == a.settingsCursor {
			cursor = "▸ "
		}
		label := fmt.Sprintf("%-*s", labelWidth, item.label)
		value := item.value
		if i == a.settingsCursor {
			label = a.theme.Cursor.Render(label)
			value = a.theme.Cursor.Render(value)
		}
		b.WriteString(cursor + label + "  " + value + "\n")
		if item.desc != "" {
			b.WriteString("    " + a.theme.Subtitle.Render(item.desc) + "\n")
		}
	}
	return b.String()
}

func (a *App) settingsKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()
	items := a.settingsItems()
	switch key {
	case "up", "k":
		if a.settingsCursor > 0 {
			a.settingsCursor--
		}
	case "down", "j":
		if a.settingsCursor < len(items)-1 {
			a.settingsCursor++
		}
	case "enter", "right", "l", " ":
		if a.settingsCursor == len(items)-1 {
			return a.confirmResetSettings()
		}
		a.changeSetting(1)
	case "left", "h":
		if a.settingsCursor == len(items)-1 {
			return nil
		}
		a.changeSetting(-1)
	case "r":
		return a.confirmResetSettings()
	case "esc", "q":
		a.screen = screenMenu
		a.settingsCursor = 0
		a.clearStatus()
	}
	return nil
}

func (a *App) confirmResetSettings() tea.Cmd {
	a.confirm = &confirmState{
		message: "确认恢复 nyaru 的默认设置？",
		onYes: func() tea.Cmd {
			a.applySettings(DefaultSettings())
			return nil
		},
	}
	return nil
}

func (a *App) changeSetting(direction int) {
	next := a.settings
	switch a.settingsCursor {
	case 0:
		next.Mode = cycleChoice([]string{"auto", "sfsu", "scoop"}, next.Mode, direction)
	case 1:
		next.ConfirmBatch = !next.ConfirmBatch
	case 2:
		next.StatusTimeout = cycleStatusTimeout(next.StatusTimeout, direction)
	default:
		return
	}
	a.applySettings(next)
}

func (a *App) applySettings(next SettingsValues) {
	next = normalizeSettings(next)
	oldMode := a.settings.Mode
	if a.saveSettings != nil {
		if err := a.saveSettings(next); err != nil {
			a.setStatus("保存设置失败: "+err.Error(), toneError)
			return
		}
	}
	a.settings = next
	a.statusTTL = next.statusTTL()
	if next.Mode != oldMode {
		a.setStatus("设置已保存；后端模式将在重启 nyaru 后生效", toneSuccess)
		return
	}
	a.setStatus("设置已保存", toneSuccess)
}

func cycleChoice(options []string, current string, direction int) string {
	index := -1
	for i, option := range options {
		if option == current {
			index = i
			break
		}
	}
	if index < 0 {
		return options[0]
	}
	index = (index + direction + len(options)) % len(options)
	return options[index]
}

func cycleStatusTimeout(current, direction int) int {
	options := []int{3, defaultStatusTimeout, 8, 0}
	for i, option := range options {
		if option == current {
			index := (i + direction + len(options)) % len(options)
			return options[index]
		}
	}
	if direction < 0 {
		return 0
	}
	return options[0]
}
