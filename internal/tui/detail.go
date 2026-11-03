package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/nyable/nyaru/internal/models"
)

// DetailContent is the data needed by the detail screen. It is deliberately
// independent from the backend so buckets and cache entries can reuse it.
type DetailContent struct {
	Title    string
	Fields   []models.DetailField
	Notes    string
	RawTitle string
	Raw      string
	URL      string
}

func DetailFromApp(app models.AppInfo) DetailContent {
	fields := make([]models.DetailField, 0, 6)
	add := func(label, value string) {
		if value != "" {
			fields = append(fields, models.DetailField{Label: label, Value: value})
		}
	}
	add("名称", app.Name)
	add("版本", app.DisplayVersion())
	add("来源", app.Origin())
	if app.Installed {
		scope := "本地"
		if app.Global {
			scope = "全局"
		}
		add("范围", scope)
	}
	add("状态", app.Status.Label())
	add("说明", app.Info)
	return DetailContent{Title: "应用详情", Fields: fields}
}

func DetailFromBucket(bucket models.BucketResult) DetailContent {
	return DetailContent{
		Title: "存储桶详情",
		Fields: []models.DetailField{
			{Label: "名称", Value: bucket.Name},
			{Label: "来源", Value: bucket.Source},
			{Label: "清单数", Value: fmt.Sprintf("%d", bucket.Manifests)},
			{Label: "更新时间", Value: bucket.Updated.DateTime},
		},
		URL: bucket.Source,
	}
}

func DetailFromCache(cache models.CacheResult) DetailContent {
	return DetailContent{
		Title: "缓存详情",
		Fields: []models.DetailField{
			{Label: "应用", Value: cache.App},
			{Label: "版本", Value: strings.Join(cache.Versions, ", ")},
			{Label: "大小", Value: cache.FormatSize},
		},
	}
}

func DetailFromOldVersion(old models.OldVersion) DetailContent {
	scope := "本地"
	if old.Global {
		scope = "全局"
	}
	fields := []models.DetailField{
		{Label: "应用", Value: old.Name},
		{Label: "当前版本", Value: old.Current},
		{Label: "范围", Value: scope},
		{Label: "旧版本数", Value: fmt.Sprintf("%d", len(old.Versions))},
		{Label: "可释放空间", Value: old.FormatSize},
	}
	var notes strings.Builder
	for _, version := range old.Versions {
		fmt.Fprintf(&notes, "• %s  %s\n", version.Version, version.FormatSize)
	}
	return DetailContent{
		Title:  "旧版本详情: " + old.Name,
		Fields: fields,
		Notes:  strings.TrimRight(notes.String(), "\n"),
	}
}

func DetailFromAppDetail(detail models.AppDetail) DetailContent {
	content := DetailContent{
		Title:  "应用详情: " + detail.Name,
		Fields: detail.Fields(),
		Notes:  detail.Notes,
		URL:    detail.Website,
	}
	if detail.Description != "" {
		content.Fields = append(content.Fields, models.DetailField{Label: "描述", Value: detail.Description})
	}
	return content
}

type DetailModel struct {
	width   int
	height  int
	content DetailContent
	view    viewport.Model
	loading bool
}

func NewDetail() *DetailModel {
	return &DetailModel{
		view: viewport.New(80, 20),
	}
}

func (d *DetailModel) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.view.Width = width
	d.view.Height = height
	d.view.SetContent(d.render())
}

func (d *DetailModel) SetContent(content DetailContent) {
	d.content = content
	d.loading = false
	d.view.SetContent(d.render())
	d.view.GotoTop()
}

func (d *DetailModel) SetLoading(title string) {
	d.loading = true
	d.content = DetailContent{Title: title}
	d.view.SetContent(d.render())
}

func (d *DetailModel) Content() DetailContent { return d.content }
func (d *DetailModel) Loading() bool          { return d.loading }

func (d *DetailModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.view, cmd = d.view.Update(msg)
	return cmd
}

func (d *DetailModel) View(theme Theme) string {
	if d.loading {
		return lipgloss.NewStyle().
			Width(d.width).
			Height(d.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(theme.Subtitle.Render("正在加载详情..."))
	}
	return d.view.View()
}

func (d *DetailModel) render() string {
	if d.loading {
		return "正在加载详情..."
	}
	var b strings.Builder
	if d.content.Title != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(d.content.Title))
		b.WriteString("\n\n")
	}
	labelWidth := 0
	for _, f := range d.content.Fields {
		if w := lipgloss.Width(f.Label); w > labelWidth {
			labelWidth = w
		}
	}
	for _, f := range d.content.Fields {
		label := fmt.Sprintf("%-*s", labelWidth, f.Label)
		b.WriteString(detailLabelStyle.Render(label) + "  " + f.Value + "\n")
	}
	if d.content.Notes != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render("备注") + "\n")
		b.WriteString(d.content.Notes + "\n")
	}
	if d.content.Raw != "" {
		title := d.content.RawTitle
		if title == "" {
			title = "输出"
		}
		width := d.width - 2
		if width < 20 {
			width = 20
		}
		b.WriteString("\n" + lipgloss.NewStyle().Bold(true).Render(title) + "\n")
		b.WriteString(ansi.HardwrapWc(d.content.Raw, width, true) + "\n")
	}
	if d.content.URL != "" {
		shown := false
		for _, f := range d.content.Fields {
			if f.Value == d.content.URL {
				shown = true
				break
			}
		}
		if !shown {
			b.WriteString("\n" + detailLabelStyle.Render("主页") + "  " + d.content.URL + "\n")
		}
	}
	return b.String()
}

// detailLabelStyle is kept package-local to avoid passing the theme into the
// render helper; it uses a neutral accent and the main theme for values.
var detailLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(
	lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#5EEAD4"},
)
