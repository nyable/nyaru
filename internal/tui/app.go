package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nyable/nyaru/internal/core"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/selfupdate"
	"github.com/nyable/nyaru/internal/utils"
)

type Start int

const (
	StartMenu Start = iota
	StartSearch
	StartInstalled
	StartUpdates
	StartBuckets
	StartCache
	StartCleanup
	StartSettings
)

type Options struct {
	Backend      core.Backend
	Start        Start
	Query        string
	Version      string
	Settings     *SettingsValues
	SaveSettings func(SettingsValues) error
}

type appScreen int

const (
	screenMenu appScreen = iota
	screenSearch
	screenList
	screenDetail
	screenOutput
	screenSettings
)

type listKind int

const (
	listNone listKind = iota
	listSearch
	listInstalled
	listUpdates
	listBuckets
	listCache
	listOldVersions
)

type tone int

const (
	toneInfo tone = iota
	toneSuccess
	toneWarning
	toneError
)

type dataMsg struct {
	kind    listKind
	query   string
	apps    []models.AppInfo
	buckets []models.BucketResult
	cache   []models.CacheResult
	old     []models.OldVersion
	err     error
}

type infoMsg struct {
	content DetailContent
	err     error
}

type execDoneMsg struct {
	err      error
	done     string
	command  string
	output   string
	exitCode int
}

type statusMsg struct {
	text string
	tone tone
}

type statusTickMsg struct{}

const statusTickInterval = 1 * time.Second

type actionItem struct {
	label       string
	desc        string
	confirm     string
	destructive bool
	onSelect    func() tea.Cmd
}

type actionState struct {
	title  string
	items  []actionItem
	cursor int
}

type confirmState struct {
	message string
	onYes   func() tea.Cmd
}

type App struct {
	backend        core.Backend
	theme          Theme
	screen         appScreen
	menu           *ListModel
	list           *ListModel
	detail         *DetailModel
	search         textinput.Model
	spinner        spinner.Model
	width          int
	height         int
	loading        bool
	loadingText    string
	status         string
	statusTone     tone
	statusAt       time.Time
	refreshNote    string
	refreshTone    tone
	help           bool
	confirm        *confirmState
	action         *actionState
	listKind       listKind
	query          string
	version        string
	settings       SettingsValues
	saveSettings   func(SettingsValues) error
	settingsCursor int
	statusTTL      time.Duration
}

func NewApp(opts Options) *App {
	m := spinner.New()
	m.Spinner = spinner.Dot

	search := textinput.New()
	search.Placeholder = "应用名或关键字"
	search.CharLimit = 120
	search.Width = 50

	settings := DefaultSettings()
	if opts.Settings != nil {
		settings = normalizeSettings(*opts.Settings)
	}
	a := &App{
		backend:      opts.Backend,
		theme:        DefaultTheme(),
		menu:         NewList("主菜单", "没有菜单项"),
		list:         NewList("", "没有内容"),
		detail:       NewDetail(),
		search:       search,
		spinner:      m,
		width:        80,
		height:       24,
		version:      opts.Version,
		settings:     settings,
		saveSettings: opts.SaveSettings,
		statusTTL:    settings.statusTTL(),
		query:        opts.Query,
	}
	a.menu.SetSelectable(false)
	a.menu.SetItems(menuItems())
	a.layout()

	switch opts.Start {
	case StartSearch:
		if opts.Query != "" {
			a.screen = screenList
			a.listKind = listSearch
			a.loading = true
			a.search.SetValue(opts.Query)
			a.list.SetTitle("搜索结果")
			a.list.SetEmpty(listEmpty(listSearch))
		} else {
			a.screen = screenSearch
			a.search.Focus()
		}
	case StartInstalled:
		a.startList(listInstalled, opts.Query)
	case StartUpdates:
		a.startList(listUpdates, "")
	case StartBuckets:
		a.startList(listBuckets, "")
	case StartCache:
		a.startList(listCache, "")
	case StartCleanup:
		a.startList(listOldVersions, "")
	case StartSettings:
		a.screen = screenSettings
	default:
		a.screen = screenMenu
	}
	return a
}

func menuItems() []Item {
	return []Item{
		MenuItem{Key: "search", Label: "搜索应用", Desc: "查找可安装的应用，支持多选安装"},
		MenuItem{Key: "installed", Label: "已安装应用", Desc: "查看、更新、卸载、锁定已安装应用"},
		MenuItem{Key: "updates", Label: "可更新应用", Desc: "检查并更新有新版本的应用"},
		MenuItem{Key: "buckets", Label: "存储桶", Desc: "查看、删除存储桶"},
		MenuItem{Key: "cache", Label: "缓存", Desc: "查看缓存占用并清理"},
		MenuItem{Key: "cleanup", Label: "清理旧版本", Desc: "删除应用的历史版本"},
		MenuItem{Key: "settings", Label: "设置", Desc: "管理 nyaru 自身的设置"},
		MenuItem{Key: "selfupdate", Label: "检查更新", Desc: "检查 nyaru 自身是否有新版本"},
		MenuItem{Key: "quit", Label: "退出", Desc: "退出 nyaru"},
	}
}

func (a *App) layout() {
	body := a.bodyHeight()
	a.menu.SetSize(a.width-2, body)
	a.list.SetSize(a.width-2, body)
	a.detail.SetSize(a.width-2, body)
	a.search.Width = max(10, a.width-8)
}

func (a *App) bodyHeight() int {
	h := a.height - 4 // header, status, footer, blank
	if h < 3 {
		h = 3
	}
	return h
}

func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{a.spinner.Tick, statusTickCmd()}
	if a.loading {
		cmds = append(cmds, a.loadCmd())
	}
	if a.screen == screenSearch {
		cmds = append(cmds, textinput.Blink)
	}
	return tea.Batch(cmds...)
}

func (a *App) loadCmd() tea.Cmd {
	kind, query := a.listKind, a.query
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		msg := dataMsg{kind: kind, query: query}
		switch kind {
		case listSearch:
			msg.apps, msg.err = a.backend.Search(ctx, query)
		case listInstalled:
			msg.apps, msg.err = a.backend.List(ctx, query)
		case listUpdates:
			msg.apps, msg.err = a.backend.Status(ctx)
		case listBuckets:
			msg.buckets, msg.err = a.backend.BucketList(ctx)
		case listCache:
			msg.cache, msg.err = a.backend.CacheList(ctx)
		case listOldVersions:
			msg.old, msg.err = a.backend.OldVersions(ctx)
		}
		return msg
	}
}

func (a *App) startList(kind listKind, query string) tea.Cmd {
	a.screen = screenList
	a.listKind = kind
	a.query = query
	a.loading = true
	a.loadingText = loadText(kind)
	a.refreshNote = ""
	a.list.ClearSelection()
	a.list.ResetFilter()
	a.list.SetItems(nil)
	a.list.SetTitle(listTitle(kind))
	a.list.SetEmpty(listEmpty(kind))
	a.clearStatus()
	a.help = false
	return tea.Batch(a.spinner.Tick, a.loadCmd())
}

func loadText(kind listKind) string {
	if kind == listOldVersions {
		return "正在扫描旧版本..."
	}
	return "正在加载..."
}

func listTitle(kind listKind) string {
	switch kind {
	case listSearch:
		return "搜索结果"
	case listInstalled:
		return "已安装应用"
	case listUpdates:
		return "可更新 / 异常应用"
	case listBuckets:
		return "存储桶"
	case listCache:
		return "缓存"
	case listOldVersions:
		return "旧版本"
	default:
		return ""
	}
}

func listEmpty(kind listKind) string {
	switch kind {
	case listSearch:
		return "没有匹配的应用"
	case listInstalled:
		return "没有已安装的应用"
	case listUpdates:
		return "没有需要处理的应用"
	case listBuckets:
		return "没有添加存储桶"
	case listCache:
		return "没有缓存"
	case listOldVersions:
		return "没有可清理的旧版本"
	default:
		return "没有内容"
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.layout()
		return a, nil

	case dataMsg:
		return a.handleData(msg)

	case infoMsg:
		if a.screen != screenDetail {
			return a, nil
		}
		a.detail.SetContent(msg.content)
		if msg.err != nil {
			a.setStatus("获取详情失败: "+msg.err.Error(), toneError)
		}
		return a, nil

	case execDoneMsg:
		a.loading = false
		a.showExecResult(msg)
		if msg.err != nil {
			failed := strings.Replace(msg.done, "完成", "失败", 1)
			a.refreshNote = failed
			a.refreshTone = toneError
			a.setStatus(fmt.Sprintf("%s: %s", failed, msg.err), toneError)
		} else {
			a.refreshNote = msg.done
			a.refreshTone = toneSuccess
			a.setStatus(msg.done+"，正在刷新列表...", toneInfo)
		}
		if a.listKind != listNone {
			a.loading = true
			a.loadingText = "正在刷新列表..."
			return a, tea.Batch(a.spinner.Tick, a.loadCmd())
		}
		return a, nil

	case statusMsg:
		a.setStatus(msg.text, msg.tone)
		return a, nil

	case statusTickMsg:
		if a.status != "" && !a.loading && a.statusTTL > 0 && time.Since(a.statusAt) >= a.statusTTL {
			a.clearStatus()
		}
		return a, statusTickCmd()

	case spinner.TickMsg:
		if a.loading {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	// Let inputs and viewport consume cursor blink / scroll messages.
	if a.screen == screenSearch && a.confirm == nil && a.action == nil && !a.help {
		var cmd tea.Cmd
		a.search, cmd = a.search.Update(msg)
		return a, cmd
	}
	if a.screen == screenDetail && a.confirm == nil && a.action == nil && !a.help {
		return a, a.detail.Update(msg)
	}
	if a.confirm == nil && a.action == nil && !a.help {
		if cmd := a.activeList().Update(msg); cmd != nil {
			return a, cmd
		}
	}
	return a, nil
}

func (a *App) handleData(msg dataMsg) (tea.Model, tea.Cmd) {
	if msg.kind != a.listKind || msg.query != a.query {
		return a, nil // stale response from a previous screen or query
	}
	a.loading = false
	a.loadingText = ""
	if a.screen != screenDetail && a.screen != screenOutput {
		a.screen = screenList
	}
	if msg.err != nil {
		if a.refreshNote != "" {
			a.setStatus(a.refreshNote+"，但刷新列表失败: "+msg.err.Error(), toneError)
		} else {
			a.setStatus("加载失败: "+msg.err.Error(), toneError)
		}
		a.list.SetItems(nil)
		a.refreshNote = ""
		return a, nil
	}
	switch msg.kind {
	case listSearch, listInstalled, listUpdates:
		items := make([]Item, 0, len(msg.apps))
		for _, app := range msg.apps {
			items = append(items, app)
		}
		a.list.SetItems(items)
		if len(items) > 0 {
			a.list.SetTitle(fmt.Sprintf("%s · %d 项", listTitle(msg.kind), len(items)))
		}
	case listBuckets:
		items := make([]Item, 0, len(msg.buckets))
		for _, bucket := range msg.buckets {
			items = append(items, bucket)
		}
		a.list.SetItems(items)
		if len(items) > 0 {
			a.list.SetTitle(fmt.Sprintf("存储桶 · %d 个", len(items)))
		}
	case listCache:
		items := make([]Item, 0, len(msg.cache))
		var total int64
		for _, cache := range msg.cache {
			items = append(items, cache)
			total += cache.Length
		}
		a.list.SetItems(items)
		a.list.SetTitle(fmt.Sprintf("缓存 · %d 个应用 · %s", len(msg.cache), utils.SizeToHumanRead(total)))
	case listOldVersions:
		items := make([]Item, 0, len(msg.old))
		var total int64
		for _, old := range msg.old {
			items = append(items, old)
			total += old.Length
		}
		a.list.SetItems(items)
		a.list.SetTitle(fmt.Sprintf("旧版本 · %d 个应用 · 可释放 %s", len(msg.old), utils.SizeToHumanRead(total)))
	}
	if len(a.list.Items()) == 0 && a.refreshNote == "" {
		a.setStatus(listEmpty(msg.kind), toneInfo)
	}
	if a.refreshNote != "" {
		text := fmt.Sprintf("%s · %d 项", a.refreshNote, a.list.TotalCount())
		if a.refreshNote != "列表已刷新" {
			text = fmt.Sprintf("%s · 列表已刷新 · %d 项", a.refreshNote, a.list.TotalCount())
		}
		a.setStatus(text, a.refreshTone)
		a.refreshNote = ""
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return a, tea.Quit
	}

	if a.help {
		if key == "?" || key == "esc" || key == "q" {
			a.help = false
		}
		return a, nil
	}

	if a.confirm != nil {
		switch key {
		case "y", "Y", "enter":
			onYes := a.confirm.onYes
			a.confirm = nil
			return a, onYes()
		case "n", "N", "esc", "q":
			a.confirm = nil
		}
		return a, nil
	}

	if a.action != nil {
		switch key {
		case "up", "k":
			if a.action.cursor > 0 {
				a.action.cursor--
			}
		case "down", "j":
			if a.action.cursor < len(a.action.items)-1 {
				a.action.cursor++
			}
		case "enter":
			if len(a.action.items) == 0 {
				a.action = nil
				return a, nil
			}
			item := a.action.items[a.action.cursor]
			a.action = nil
			if item.confirm != "" && (a.settings.ConfirmBatch || item.destructive) {
				a.confirm = &confirmState{message: item.confirm, onYes: item.onSelect}
				return a, nil
			}
			return a, item.onSelect()
		case "esc", "q":
			a.action = nil
		}
		return a, nil
	}

	switch key {
	case "?":
		a.help = true
		a.clearStatus()
		return a, nil
	}

	switch a.screen {
	case screenMenu:
		action, cmd := a.menu.HandleKey(msg)
		switch action {
		case actEnter:
			if item := a.menu.CurrentItem(); item != nil {
				return a, a.activateMenu(item)
			}
		case actBack:
			return a, tea.Quit
		}
		return a, cmd

	case screenSearch:
		switch key {
		case "esc":
			a.screen = screenMenu
			a.search.Blur()
			a.clearStatus()
			return a, nil
		case "enter":
			query := strings.TrimSpace(a.search.Value())
			if query == "" {
				a.setStatus("请输入搜索关键字", toneWarning)
				return a, nil
			}
			a.search.Blur()
			a.query = query
			return a, a.startList(listSearch, query)
		}
		var cmd tea.Cmd
		a.search, cmd = a.search.Update(msg)
		return a, cmd

	case screenList:
		if key == "/" {
			a.clearStatus()
		}
		action, cmd := a.list.HandleKey(msg)
		switch action {
		case actEnter:
			return a, a.openActions()
		case actInfo:
			return a, a.openDetail()
		case actOpen:
			return a, a.openURL()
		case actRefresh:
			a.loading = true
			a.loadingText = "正在刷新列表..."
			a.refreshNote = "列表已刷新"
			a.refreshTone = toneSuccess
			a.setStatus("正在刷新列表...", toneInfo)
			return a, tea.Batch(a.spinner.Tick, a.loadCmd())
		case actBack:
			if a.list.Filtering() {
				return a, cmd
			}
			return a, a.backFromList()
		}
		return a, cmd

	case screenOutput:
		switch key {
		case "esc", "enter", "q", "backspace":
			return a, a.closeOutput()
		}
		return a, a.detail.Update(msg)

	case screenSettings:
		return a, a.settingsKey(msg)

	case screenDetail:
		switch key {
		case "esc", "q", "backspace":
			a.screen = screenList
			a.clearStatus()
			return a, nil
		case "o":
			return a, a.openURL()
		}
		return a, a.detail.Update(msg)
	}
	return a, nil
}

func (a *App) closeOutput() tea.Cmd {
	a.clearStatus()
	if a.listKind != listNone {
		a.screen = screenList
	} else {
		a.screen = screenMenu
	}
	return nil
}

func (a *App) activateMenu(item Item) tea.Cmd {
	switch item.ID() {
	case "search":
		a.screen = screenSearch
		a.search.Focus()
		a.clearStatus()
		return textinput.Blink
	case "installed":
		return a.startList(listInstalled, "")
	case "updates":
		return a.startList(listUpdates, "")
	case "buckets":
		return a.startList(listBuckets, "")
	case "cache":
		return a.startList(listCache, "")
	case "cleanup":
		return a.startList(listOldVersions, "")
	case "settings":
		a.screen = screenSettings
		a.settingsCursor = 0
		a.clearStatus()
		return nil
	case "selfupdate":
		a.setStatus("正在检查 nyaru 更新...", toneInfo)
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			status, err := selfupdate.Check(ctx, a.version)
			if err != nil {
				return statusMsg{text: "检查更新失败: " + err.Error(), tone: toneError}
			}
			if status.UpdateAvailable {
				return statusMsg{
					text: fmt.Sprintf("发现新版本 v%s，运行 nyaru self-update 更新", status.Latest),
					tone: toneWarning,
				}
			}
			return statusMsg{text: "nyaru 已是最新版本", tone: toneSuccess}
		}
	case "quit":
		return tea.Quit
	}
	return nil
}

func (a *App) activeList() *ListModel {
	if a.screen == screenMenu {
		return a.menu
	}
	return a.list
}

func (a *App) backFromList() tea.Cmd {
	a.loading = false
	a.loadingText = ""
	a.refreshNote = ""
	a.clearStatus()
	if a.listKind == listSearch {
		a.screen = screenSearch
		a.listKind = listNone
		a.search.Focus()
		return textinput.Blink
	}
	a.screen = screenMenu
	a.listKind = listNone
	return nil
}

func (a *App) selectedOrCurrent() []Item {
	items := a.list.SelectedItems()
	if len(items) > 0 {
		return items
	}
	if item := a.list.CurrentItem(); item != nil {
		return []Item{item}
	}
	return nil
}

type appGroup struct {
	label  string
	names  []string
	global bool
}

// splitAppGroups separates local and global apps so a mixed selection is not
// sent to scoop with the wrong --global flag.
func splitAppGroups(items []Item) []appGroup {
	var local, global []string
	for _, item := range items {
		app, ok := item.(models.AppInfo)
		if !ok {
			continue
		}
		if app.Global {
			global = append(global, app.Name)
		} else {
			local = append(local, app.Name)
		}
	}
	groups := make([]appGroup, 0, 2)
	if len(local) > 0 {
		groups = append(groups, appGroup{label: "本地", names: local})
	}
	if len(global) > 0 {
		groups = append(groups, appGroup{label: "全局", names: global, global: true})
	}
	return groups
}

// lookupName uses the plain name for installed apps and the bucket-qualified
// name for search results, so info works with both sfsu and scoop.
func lookupName(app models.AppInfo) string {
	if app.Installed || app.Name == "" {
		return app.Name
	}
	return app.FullName()
}

func installNames(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if app, ok := item.(models.AppInfo); ok {
			out = append(out, app.FullName())
		}
	}
	return out
}

func withoutSelf(names []string) ([]string, bool) {
	out := make([]string, 0, len(names))
	skipped := false
	for _, name := range names {
		if strings.EqualFold(name, "nyaru") {
			skipped = true
			continue
		}
		out = append(out, name)
	}
	return out, skipped
}

func cacheNames(items []Item) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if cache, ok := item.(models.CacheResult); ok {
			out = append(out, cache.App)
		}
	}
	return out
}

// splitOldVersionGroups keeps local and global cleanup commands separate so
// the --global flag matches the app that is actually being cleaned.
func splitOldVersionGroups(items []Item) []appGroup {
	var local, global []string
	for _, item := range items {
		old, ok := item.(models.OldVersion)
		if !ok {
			continue
		}
		if old.Global {
			global = append(global, old.Name)
		} else {
			local = append(local, old.Name)
		}
	}
	groups := make([]appGroup, 0, 2)
	if len(local) > 0 {
		groups = append(groups, appGroup{label: "本地", names: local})
	}
	if len(global) > 0 {
		groups = append(groups, appGroup{label: "全局", names: global, global: true})
	}
	return groups
}

func oldVersionSize(items []Item, global bool) int64 {
	var total int64
	for _, item := range items {
		if old, ok := item.(models.OldVersion); ok && old.Global == global {
			total += old.Length
		}
	}
	return total
}

func appRunning(verb string, count int) string {
	return fmt.Sprintf("正在%s %d 个应用...", verb, count)
}

func appDone(verb string, count int) string {
	return fmt.Sprintf("%s完成（%d 个应用）", verb, count)
}

func (a *App) openActions() tea.Cmd {
	items := a.selectedOrCurrent()
	if len(items) == 0 {
		return nil
	}
	a.clearStatus()
	var actions []actionItem
	switch a.listKind {
	case listSearch:
		names := installNames(items)
		if len(names) == 0 {
			return nil
		}
		actions = []actionItem{
			{
				label:   fmt.Sprintf("安装选中的 %d 个应用", len(names)),
				desc:    strings.Join(names, ", "),
				confirm: fmt.Sprintf("确认安装这 %d 个应用？\n%s", len(names), strings.Join(names, ", ")),
				onSelect: func() tea.Cmd {
					return a.execCommand(
						a.backend.Install(names, core.InstallOptions{}),
						appRunning("安装", len(names)),
						appDone("安装", len(names)),
					)
				},
			},
			{label: "查看详情", desc: "查看当前应用的信息", onSelect: a.openDetail},
		}
	case listInstalled:
		groups := splitAppGroups(items)
		multi := len(groups) > 1
		actions = []actionItem{}
		skippedSelf := false
		for _, group := range groups {
			if len(group.names) == 0 {
				continue
			}
			scope := ""
			if multi || group.global {
				scope = group.label
			}
			updateApps, skipped := withoutSelf(group.names)
			skippedSelf = skippedSelf || skipped
			if len(updateApps) > 0 {
				names := updateApps
				actions = append(actions, actionItem{
					label:   fmt.Sprintf("更新选中的 %d 个%s应用", len(names), scope),
					desc:    strings.Join(names, ", "),
					confirm: fmt.Sprintf("确认更新这 %d 个%s应用？\n%s", len(names), scope, strings.Join(names, ", ")),
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Update(names, core.UpdateOptions{Global: group.global}),
							appRunning("更新", len(names)),
							appDone("更新", len(names)),
						)
					},
				})
			}
			names := group.names
			actions = append(actions,
				actionItem{
					label:       fmt.Sprintf("卸载选中的 %d 个%s应用", len(names), scope),
					desc:        strings.Join(names, ", "),
					confirm:     fmt.Sprintf("确认卸载这 %d 个%s应用？此操作不可撤销。\n%s", len(names), scope, strings.Join(names, ", ")),
					destructive: true,
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Uninstall(names, core.UninstallOptions{Global: group.global}),
							appRunning("卸载", len(names)),
							appDone("卸载", len(names)),
						)
					},
				},
				actionItem{
					label:   fmt.Sprintf("锁定选中的 %d 个%s应用", len(names), scope),
					desc:    "锁定后 scoop update 将跳过这些应用",
					confirm: fmt.Sprintf("确认锁定这 %d 个%s应用？\n%s", len(names), scope, strings.Join(names, ", ")),
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Hold(names, core.HoldOptions{Global: group.global}),
							appRunning("锁定", len(names)),
							appDone("锁定", len(names)),
						)
					},
				},
				actionItem{
					label:   fmt.Sprintf("解锁选中的 %d 个%s应用", len(names), scope),
					desc:    "恢复这些应用的自动更新",
					confirm: fmt.Sprintf("确认解锁这 %d 个%s应用？\n%s", len(names), scope, strings.Join(names, ", ")),
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Unhold(names, core.HoldOptions{Global: group.global}),
							appRunning("解锁", len(names)),
							appDone("解锁", len(names)),
						)
					},
				},
				actionItem{
					label:   fmt.Sprintf("清理选中的 %d 个%s应用的旧版本", len(names), scope),
					desc:    "删除旧版本目录，保留当前版本",
					confirm: fmt.Sprintf("确认清理这 %d 个%s应用的旧版本？\n%s", len(names), scope, strings.Join(names, ", ")),
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Cleanup(names, core.CleanupOptions{Global: group.global}),
							fmt.Sprintf("正在清理旧版本（%d 个应用）...", len(names)),
							fmt.Sprintf("旧版本清理完成（%d 个应用）", len(names)),
						)
					},
				},
			)
		}
		if skippedSelf {
			a.setStatus("nyaru 自身请使用 nyaru self-update 更新，已从批量更新中排除", toneWarning)
		}
		actions = append(actions, actionItem{label: "查看详情", desc: "查看当前应用的信息", onSelect: a.openDetail})
	case listUpdates:
		groups := splitAppGroups(items)
		multi := len(groups) > 1
		actions = []actionItem{}
		skippedSelf := false
		for _, group := range groups {
			if len(group.names) == 0 {
				continue
			}
			scope := ""
			if multi || group.global {
				scope = group.label
			}
			updateApps, skipped := withoutSelf(group.names)
			skippedSelf = skippedSelf || skipped
			if len(updateApps) > 0 {
				names := updateApps
				actions = append(actions, actionItem{
					label:   fmt.Sprintf("更新选中的 %d 个%s应用", len(names), scope),
					desc:    strings.Join(names, ", "),
					confirm: fmt.Sprintf("确认更新这 %d 个%s应用？\n%s", len(names), scope, strings.Join(names, ", ")),
					onSelect: func() tea.Cmd {
						return a.execCommand(
							a.backend.Update(names, core.UpdateOptions{Global: group.global}),
							appRunning("更新", len(names)),
							appDone("更新", len(names)),
						)
					},
				})
			}
			names := group.names
			actions = append(actions, actionItem{
				label:   fmt.Sprintf("锁定选中的 %d 个%s应用", len(names), scope),
				desc:    "锁定后不再提示更新",
				confirm: fmt.Sprintf("确认锁定这 %d 个%s应用？\n%s", len(names), scope, strings.Join(names, ", ")),
				onSelect: func() tea.Cmd {
					return a.execCommand(
						a.backend.Hold(names, core.HoldOptions{Global: group.global}),
						appRunning("锁定", len(names)),
						appDone("锁定", len(names)),
					)
				},
			})
		}
		if skippedSelf {
			a.setStatus("nyaru 自身请使用 nyaru self-update 更新，已从批量更新中排除", toneWarning)
		}
		actions = append(actions, actionItem{label: "查看详情", desc: "查看当前应用的信息", onSelect: a.openDetail})
	case listOldVersions:
		groups := splitOldVersionGroups(items)
		multi := len(groups) > 1
		actions = []actionItem{}
		for _, group := range groups {
			if len(group.names) == 0 {
				continue
			}
			scope := ""
			if multi || group.global {
				scope = group.label
			}
			names := group.names
			size := utils.SizeToHumanRead(oldVersionSize(items, group.global))
			actions = append(actions, actionItem{
				label:   fmt.Sprintf("清理选中的 %d 个%s应用的旧版本", len(names), scope),
				desc:    strings.Join(names, ", ") + "\n可释放 " + size,
				confirm: fmt.Sprintf("确认清理这 %d 个%s应用的旧版本？可释放 %s。\n%s", len(names), scope, size, strings.Join(names, ", ")),
				onSelect: func() tea.Cmd {
					return a.execCommand(
						a.backend.Cleanup(names, core.CleanupOptions{Global: group.global}),
						fmt.Sprintf("正在清理旧版本（%d 个应用）...", len(names)),
						fmt.Sprintf("旧版本清理完成（%d 个应用）", len(names)),
					)
				},
			})
		}
		actions = append(actions, actionItem{label: "查看详情", desc: "查看旧版本及占用空间", onSelect: a.openDetail})
	case listBuckets:
		current := a.list.CurrentItem()
		if current == nil {
			return nil
		}
		bucket, ok := current.(models.BucketResult)
		if !ok {
			return nil
		}
		actions = []actionItem{
			{
				label:       "删除当前存储桶",
				desc:        bucket.Name,
				confirm:     fmt.Sprintf("确认删除存储桶 %s？", bucket.Name),
				destructive: true,
				onSelect: func() tea.Cmd {
					return a.execCommand(
						a.backend.BucketRemove(bucket.Name),
						"正在删除存储桶 "+bucket.Name+"...",
						"存储桶删除完成（"+bucket.Name+"）",
					)
				},
			},
			{
				label: "查看添加命令",
				desc:  "显示重新添加该存储桶的命令",
				onSelect: func() tea.Cmd {
					return func() tea.Msg {
						return statusMsg{
							text: fmt.Sprintf("scoop bucket add %s %s", bucket.Name, bucket.Source),
							tone: toneInfo,
						}
					}
				},
			},
			{label: "查看详情", desc: "查看存储桶信息", onSelect: a.openDetail},
		}
	case listCache:
		names := cacheNames(items)
		if len(names) == 0 {
			return nil
		}
		all := len(names) == a.list.TotalCount() && len(names) > 0
		actions = []actionItem{
			{
				label:       fmt.Sprintf("删除选中的 %d 个应用缓存", len(names)),
				desc:        strings.Join(names, ", ") + "\n注意：会删除这些应用的全部缓存版本",
				confirm:     fmt.Sprintf("确认删除这 %d 个应用的全部缓存？\n%s", len(names), strings.Join(names, ", ")),
				destructive: true,
				onSelect: func() tea.Cmd {
					return a.execCommand(
						a.backend.CacheRemove(names, all),
						fmt.Sprintf("正在清理缓存（%d 个应用）...", len(names)),
						fmt.Sprintf("缓存清理完成（%d 个应用）", len(names)),
					)
				},
			},
			{label: "查看详情", desc: "查看缓存信息", onSelect: a.openDetail},
		}
	}
	if len(actions) == 0 {
		return nil
	}
	a.action = &actionState{title: "选择操作", items: actions}
	return nil
}

func (a *App) execCommand(cmd core.Command, running, done string) tea.Cmd {
	a.loading = true
	a.loadingText = running
	a.refreshNote = ""
	a.setStatus(running, toneInfo)
	c := cmd
	banner := fmt.Sprintf("==> %s\n    命令: %s\n", running, c.String())
	exec := newCommandWithBanner(c.Exec(), banner)
	return tea.Exec(exec, func(err error) tea.Msg {
		return execDoneMsg{
			err:      err,
			done:     done,
			command:  c.String(),
			output:   exec.Output(),
			exitCode: commandExitCode(err),
		}
	})
}

func (a *App) showExecResult(msg execDoneMsg) {
	outcome := msg.done
	if msg.err != nil {
		outcome = strings.Replace(msg.done, "完成", "失败", 1)
	}
	fields := []models.DetailField{{Label: "命令", Value: msg.command}}
	if msg.exitCode >= 0 {
		fields = append(fields, models.DetailField{Label: "退出码", Value: fmt.Sprintf("%d", msg.exitCode)})
	}
	if msg.err != nil {
		fields = append(fields, models.DetailField{Label: "错误", Value: msg.err.Error()})
	}
	output := msg.output
	if strings.TrimSpace(output) == "" {
		output = "（命令没有输出）"
	}
	a.detail.SetContent(DetailContent{
		Title:    "执行结果: " + outcome,
		Fields:   fields,
		RawTitle: "输出",
		Raw:      output,
	})
	a.screen = screenOutput
}

func (a *App) openDetail() tea.Cmd {
	item := a.list.CurrentItem()
	if item == nil {
		return nil
	}
	a.clearStatus()
	a.screen = screenDetail
	a.detail.SetLoading(item.Title())
	switch v := item.(type) {
	case models.BucketResult:
		a.detail.SetContent(DetailFromBucket(v))
		return nil
	case models.CacheResult:
		a.detail.SetContent(DetailFromCache(v))
		return nil
	case models.OldVersion:
		a.detail.SetContent(DetailFromOldVersion(v))
		return nil
	case models.AppInfo:
		name := lookupName(v)
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			detail, err := a.backend.Info(ctx, name)
			if err != nil {
				return infoMsg{content: DetailFromApp(v), err: err}
			}
			return infoMsg{content: DetailFromAppDetail(detail)}
		}
	}
	return nil
}

func (a *App) openURL() tea.Cmd {
	url := ""
	if a.screen == screenDetail {
		url = a.detail.Content().URL
	}
	if url == "" {
		if item := a.list.CurrentItem(); item != nil {
			switch v := item.(type) {
			case models.BucketResult:
				url = v.Source
			case models.AppInfo:
				name := lookupName(v)
				return func() tea.Msg {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
					defer cancel()
					detail, err := a.backend.Info(ctx, name)
					if err != nil {
						return statusMsg{text: "获取主页失败: " + err.Error(), tone: toneError}
					}
					if detail.Website == "" {
						return statusMsg{text: "当前应用没有配置主页", tone: toneWarning}
					}
					if err := utils.OpenBrowser(detail.Website); err != nil {
						return statusMsg{text: "打开浏览器失败: " + err.Error(), tone: toneError}
					}
					return statusMsg{text: "已在浏览器中打开: " + detail.Website, tone: toneSuccess}
				}
			}
		}
	}
	if url == "" {
		a.setStatus("当前项目没有可打开的主页", toneWarning)
		return nil
	}
	if err := utils.OpenBrowser(url); err != nil {
		a.setStatus("打开浏览器失败: "+err.Error(), toneError)
		return nil
	}
	a.setStatus("已在浏览器中打开: "+url, toneSuccess)
	return nil
}

func statusTickCmd() tea.Cmd {
	return tea.Tick(statusTickInterval, func(time.Time) tea.Msg {
		return statusTickMsg{}
	})
}

func (a *App) setStatus(text string, t tone) {
	a.status = text
	a.statusTone = t
	a.statusAt = time.Now()
}

func (a *App) clearStatus() {
	a.status = ""
	a.statusTone = toneInfo
	a.statusAt = time.Time{}
}

func (a *App) View() string {
	if a.width <= 0 {
		a.width = 80
	}
	if a.height <= 0 {
		a.height = 24
	}
	header := a.headerView()
	body := a.bodyView()
	status := a.statusView()
	footer := a.footerView()
	return lipgloss.NewStyle().Padding(0, 1).Render(
		lipgloss.JoinVertical(lipgloss.Left, header, body, status, footer),
	)
}

func (a *App) headerView() string {
	left := a.theme.Title.Render("nyaru")
	if a.version != "" {
		left += " " + a.theme.Subtitle.Render("v"+a.version)
	}
	left += " " + a.theme.Subtitle.Render("["+a.backend.Name()+"]")

	title := ""
	right := ""
	switch a.screen {
	case screenMenu:
		title = "主菜单"
	case screenSearch:
		title = "搜索应用"
	case screenList:
		title = a.list.Title()
		right = fmt.Sprintf("已选 %d / 共 %d", a.list.SelectedCount(), a.list.TotalCount())
	case screenDetail, screenOutput:
		title = a.detail.Content().Title
	case screenSettings:
		title = "设置"
	}
	center := a.theme.Subtitle.Render(title)
	if right != "" {
		center += "  " + a.theme.Help.Render(right)
	}
	gap := a.width - lipgloss.Width(left) - lipgloss.Width(center) - 2
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + center
}

func (a *App) bodyView() string {
	if a.help {
		return a.helpView()
	}
	if a.confirm != nil {
		return a.confirmView()
	}
	if a.action != nil {
		return a.actionView()
	}
	if a.loading && a.screen != screenOutput {
		text := a.loadingText
		if text == "" {
			text = "正在加载..."
		}
		return lipgloss.NewStyle().
			Width(a.width-2).
			Height(a.bodyHeight()).
			Align(lipgloss.Center, lipgloss.Center).
			Render(a.spinner.View() + " " + text)
	}
	switch a.screen {
	case screenMenu:
		return a.menu.View(a.theme)
	case screenSearch:
		return a.searchView()
	case screenList:
		return a.list.View(a.theme)
	case screenDetail, screenOutput:
		return a.detail.View(a.theme)
	case screenSettings:
		return a.settingsView()
	}
	return ""
}

func (a *App) searchView() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		a.theme.Subtitle.Render("输入关键字搜索可安装的应用"),
		"",
		a.search.View(),
	)
}

func (a *App) statusView() string {
	text := a.status
	if text == "" {
		return ""
	}
	switch a.statusTone {
	case toneSuccess:
		return a.theme.SuccessTxt.Render("✓ " + text)
	case toneWarning:
		return a.theme.WarnText.Render("⚠ " + text)
	case toneError:
		return a.theme.ErrorText.Render("✗ " + text)
	default:
		return a.theme.Status.Render("ℹ " + text)
	}
}

func (a *App) footerView() string {
	var keys string
	switch {
	case a.help:
		keys = "? 关闭帮助"
	case a.confirm != nil:
		keys = "y / Enter 确认 · n / Esc 取消"
	case a.action != nil:
		keys = "↑↓ 移动 · Enter 选择 · Esc 取消"
	case a.screen == screenMenu:
		keys = "↑↓ 移动 · Enter 选择 · ? 帮助 · q 退出"
	case a.screen == screenSearch:
		keys = "Enter 搜索 · Esc 返回"
	case a.screen == screenOutput:
		keys = "↑↓ 滚动 · Enter / Esc 关闭"
	case a.screen == screenSettings:
		keys = "↑↓ 移动 · Enter / ←→ 修改 · r 恢复默认 · Esc 返回"
	case a.screen == screenDetail:
		keys = "↑↓ 滚动 · o 打开主页 · Esc 返回"
	case a.screen == screenList && a.list.Filtering():
		keys = "输入关键字 · Enter 应用 · Esc 清除"
	default:
		keys = "↑↓ 移动 · Space 选择 · a 全选 · Enter 操作 · i 详情 · / 过滤 · r 刷新 · ? 帮助 · Esc 返回"
	}
	return a.theme.Help.Render(keys)
}

func (a *App) actionView() string {
	var b strings.Builder
	b.WriteString(a.theme.Title.Render(a.action.title))
	b.WriteString("\n\n")
	for i, item := range a.action.items {
		cursor := "  "
		if i == a.action.cursor {
			cursor = "▸ "
		}
		label := item.label
		if i == a.action.cursor {
			label = a.theme.Cursor.Render(label)
		}
		b.WriteString(cursor + label + "\n")
		if item.desc != "" {
			b.WriteString("    " + a.theme.Subtitle.Render(item.desc) + "\n")
		}
	}
	return a.theme.Modal.Render(b.String())
}

func (a *App) confirmView() string {
	content := a.theme.Title.Render("请确认") + "\n\n" + a.confirm.message
	return a.theme.Modal.Render(content)
}

func (a *App) helpView() string {
	lines := []string{
		a.theme.Title.Render("快捷键"),
		"",
		"↑ / ↓ / j / k   移动",
		"Space           选中 / 取消",
		"a / A           全选当前结果 / 全选全部",
		"c               清空选择",
		"Enter           执行主操作",
		"i               查看详情",
		"/               过滤",
		"o               打开主页 / 源地址",
		"r               刷新",
		"?               打开 / 关闭帮助",
		"Esc             返回",
		"q               返回上一级",
		"Ctrl+C          退出",
	}
	return a.theme.Modal.Render(strings.Join(lines, "\n"))
}

func Run(opts Options) error {
	p := tea.NewProgram(NewApp(opts), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
