package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
)

type fakeBackend struct{}

func (fakeBackend) Name() string { return "test" }
func (fakeBackend) Search(context.Context, string) ([]models.AppInfo, error) {
	return nil, nil
}
func (fakeBackend) List(context.Context, string) ([]models.AppInfo, error) { return nil, nil }
func (fakeBackend) Status(context.Context) ([]models.AppInfo, error)       { return nil, nil }
func (fakeBackend) Info(context.Context, string) (models.AppDetail, error) {
	return models.AppDetail{}, nil
}
func (fakeBackend) BucketList(context.Context) ([]models.BucketResult, error) { return nil, nil }
func (fakeBackend) CacheList(context.Context) ([]models.CacheResult, error)   { return nil, nil }
func (fakeBackend) OldVersions(context.Context) ([]models.OldVersion, error)  { return nil, nil }
func (fakeBackend) Install([]string, core.InstallOptions) core.Command {
	return core.Command{Name: "scoop"}
}
func (fakeBackend) Uninstall([]string, core.UninstallOptions) core.Command {
	return core.Command{Name: "scoop"}
}
func (fakeBackend) Update([]string, core.UpdateOptions) core.Command {
	return core.Command{Name: "scoop"}
}
func (fakeBackend) Hold([]string, core.HoldOptions) core.Command { return core.Command{Name: "scoop"} }
func (fakeBackend) Unhold([]string, core.HoldOptions) core.Command {
	return core.Command{Name: "scoop"}
}
func (fakeBackend) BucketAdd(string, string) core.Command   { return core.Command{Name: "scoop"} }
func (fakeBackend) BucketRemove(string) core.Command        { return core.Command{Name: "scoop"} }
func (fakeBackend) CacheRemove([]string, bool) core.Command { return core.Command{Name: "scoop"} }
func (fakeBackend) Cleanup([]string, core.CleanupOptions) core.Command {
	return core.Command{Name: "scoop"}
}

func TestAppMenuView(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	_, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := app.View()
	if view == "" {
		t.Fatal("expected non-empty menu view")
	}
	if !strings.Contains(view, "主菜单") {
		t.Fatalf("menu title missing from view:\n%s", view)
	}
}

func TestAppSearchEscReturnsToMenu(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartSearch, Version: "test"})
	_, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.screen != screenMenu {
		t.Fatalf("expected menu after Esc, got %v", app.screen)
	}
}

func TestAppInstalledOpensActionMenu(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listInstalled
	app.list.SetItems([]Item{models.AppInfo{Name: "git", Installed: true}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.action == nil {
		t.Fatal("expected action menu")
	}
	if len(app.action.items) == 0 {
		t.Fatal("expected actions")
	}
}

func TestInstalledActionsSplitLocalAndGlobal(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listInstalled
	app.list.SetItems([]Item{
		models.AppInfo{Name: "git", Installed: true},
		models.AppInfo{Name: "7zip", Installed: true, Global: true},
	})
	app.list.selectAll()
	app.openActions()
	if app.action == nil {
		t.Fatal("expected action menu")
	}
	var labels []string
	for _, item := range app.action.items {
		labels = append(labels, item.label)
	}
	joined := strings.Join(labels, "\n")
	if !strings.Contains(joined, "本地") || !strings.Contains(joined, "全局") {
		t.Fatalf("expected separate local and global actions:\n%s", joined)
	}
}

func TestOldVersionActionsSplitLocalAndGlobal(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listOldVersions
	app.list.SetItems([]Item{
		models.OldVersion{Name: "git", Current: "2.47.0", Length: 1048576, FormatSize: "1.0 MB", Versions: []models.OldVersionEntry{{Version: "2.40.0"}}},
		models.OldVersion{Name: "7zip", Current: "24.09", Global: true, Length: 2097152, FormatSize: "2.0 MB", Versions: []models.OldVersionEntry{{Version: "24.08"}}},
	})
	app.list.selectAll()
	app.openActions()
	if app.action == nil {
		t.Fatal("expected cleanup action menu")
	}
	var labels []string
	for _, item := range app.action.items {
		labels = append(labels, item.label)
	}
	joined := strings.Join(labels, "\n")
	if !strings.Contains(joined, "本地") || !strings.Contains(joined, "全局") {
		t.Fatalf("expected separate local and global cleanup actions:\n%s", joined)
	}
	if !strings.Contains(joined, "旧版本") {
		t.Fatalf("expected cleanup actions:\n%s", joined)
	}
}

func TestExecDoneRefreshesListWithFeedback(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listInstalled
	app.loading = false

	_, cmd := app.Update(execDoneMsg{done: "卸载完成（1 个应用）"})
	if cmd == nil {
		t.Fatal("expected reload command after execution")
	}
	if !app.loading || app.loadingText != "正在刷新列表..." {
		t.Fatalf("expected refresh progress, loading=%v text=%q", app.loading, app.loadingText)
	}

	_, _ = app.Update(dataMsg{
		kind: listInstalled,
		apps: []models.AppInfo{{Name: "git", Installed: true}},
	})
	if app.loading {
		t.Fatal("expected loading to finish")
	}
	if !strings.Contains(app.status, "卸载完成") || !strings.Contains(app.status, "列表已刷新") {
		t.Fatalf("missing completion feedback: %q", app.status)
	}
}

func TestOldVersionListShowsSizes(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listOldVersions
	_, _ = app.Update(dataMsg{
		kind: listOldVersions,
		old: []models.OldVersion{{
			Name:       "git",
			Current:    "2.47.0",
			Length:     3145728,
			FormatSize: "3.0 MB",
			Versions: []models.OldVersionEntry{
				{Version: "2.40.0", FormatSize: "1.0 MB"},
				{Version: "2.45.0", FormatSize: "2.0 MB"},
			},
		}},
	})
	if !strings.Contains(app.list.Title(), "3.0 MB") {
		t.Fatalf("title missing total size: %q", app.list.Title())
	}
	view := app.View()
	if !strings.Contains(view, "3.0 MB") || !strings.Contains(view, "旧版本") {
		t.Fatalf("list view missing old-version details:\n%s", view)
	}
}

func TestExecResultShowsCommandAndOutput(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.screen = screenList
	app.listKind = listInstalled
	_, _ = app.Update(execDoneMsg{
		err:      errors.New("exit status 1"),
		done:     "更新完成（1 个应用）",
		command:  "scoop update git",
		output:   "ERROR Couldn't find manifest for 'git'",
		exitCode: 1,
	})
	if app.screen != screenOutput {
		t.Fatalf("expected output screen, got %v", app.screen)
	}
	view := app.View()
	for _, want := range []string{"执行结果", "更新失败", "scoop update git", "退出码", "1", "Couldn't find manifest"} {
		if !strings.Contains(view, want) {
			t.Fatalf("result view missing %q:\n%s", want, view)
		}
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.screen != screenList {
		t.Fatalf("expected list after closing output, got %v", app.screen)
	}
	if app.status != "" {
		t.Fatalf("expected status to clear when closing output, got %q", app.status)
	}
}

func TestSettingsScreenChangesValues(t *testing.T) {
	var saved SettingsValues
	settings := SettingsValues{Mode: "auto", ConfirmBatch: true, StatusTimeout: 5}
	app := NewApp(Options{
		Backend:      fakeBackend{},
		Start:        StartSettings,
		Version:      "test",
		Settings:     &settings,
		SaveSettings: func(v SettingsValues) error { saved = v; return nil },
	})
	if app.screen != screenSettings {
		t.Fatalf("expected settings screen, got %v", app.screen)
	}
	view := app.View()
	for _, want := range []string{"设置", "后端模式", "批量操作确认", "状态提示时长"} {
		if !strings.Contains(view, want) {
			t.Fatalf("settings view missing %q:\n%s", want, view)
		}
	}

	app.settingsCursor = 1
	_, _ = app.Update(keyRune(' '))
	if saved.ConfirmBatch {
		t.Fatal("expected confirm_batch to be saved as false")
	}
	if app.settings.ConfirmBatch {
		t.Fatal("expected confirm_batch to update in memory")
	}

	app.settingsCursor = 0
	_, _ = app.Update(keyRune('l'))
	if saved.Mode != "sfsu" {
		t.Fatalf("expected mode sfsu, got %q", saved.Mode)
	}
	if !strings.Contains(app.status, "设置已保存") {
		t.Fatalf("missing save feedback: %q", app.status)
	}
}

func TestSettingsReset(t *testing.T) {
	var saved SettingsValues
	settings := SettingsValues{Mode: "scoop", ConfirmBatch: false, StatusTimeout: 0}
	app := NewApp(Options{
		Backend:      fakeBackend{},
		Start:        StartSettings,
		Version:      "test",
		Settings:     &settings,
		SaveSettings: func(v SettingsValues) error { saved = v; return nil },
	})
	app.settingsCursor = 3
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.confirm == nil {
		t.Fatal("expected reset confirmation")
	}
	_, _ = app.Update(keyRune('y'))
	if saved.Mode != "auto" || !saved.ConfirmBatch || saved.StatusTimeout != defaultStatusTimeout {
		t.Fatalf("unexpected defaults: %+v", saved)
	}
}

func TestStatusExpires(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	app.setStatus("更新完成", toneSuccess)
	app.statusAt = time.Now().Add(-app.statusTTL - time.Second)
	_, cmd := app.Update(statusTickMsg{})
	if app.status != "" {
		t.Fatalf("expected status to expire, got %q", app.status)
	}
	if cmd == nil {
		t.Fatal("expected status ticker to continue")
	}
}

func TestCaptureBufferTruncates(t *testing.T) {
	capture := newCaptureBuffer(5)
	if _, err := capture.Write([]byte("abcdef")); err != nil {
		t.Fatal(err)
	}
	out := capture.String()
	if !strings.Contains(out, "abcde") || !strings.Contains(out, "已截断") {
		t.Fatalf("unexpected capture: %q", out)
	}
}

func TestAppViewsAtSmallSize(t *testing.T) {
	app := NewApp(Options{Backend: fakeBackend{}, Start: StartMenu, Version: "test"})
	_, _ = app.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	if app.View() == "" {
		t.Fatal("menu view is empty")
	}

	app.screen = screenList
	app.listKind = listInstalled
	app.list.SetItems([]Item{models.AppInfo{Name: "git", CurrentVersion: "2.40.0", AvailableVersion: "2.47.0", Installed: true}})
	if app.View() == "" {
		t.Fatal("list view is empty")
	}

	app.action = &actionState{title: "操作", items: []actionItem{{label: "更新", desc: "git"}}}
	if app.View() == "" {
		t.Fatal("action view is empty")
	}

	app.action = nil
	app.confirm = &confirmState{message: "确认？", onYes: func() tea.Cmd { return nil }}
	if app.View() == "" {
		t.Fatal("confirm view is empty")
	}

	app.confirm = nil
	app.help = true
	if app.View() == "" {
		t.Fatal("help view is empty")
	}
}
