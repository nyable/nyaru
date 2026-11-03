package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type listAction int

const (
	actNone listAction = iota
	actEnter
	actBack
	actInfo
	actOpen
	actRefresh
)

// ListModel is a custom list with filtering, multi-selection and consistent
// two-line rows. It deliberately avoids bubbles/list so every screen shares
// the same layout and key handling.
type ListModel struct {
	title      string
	empty      string
	items      []Item
	visible    []int
	cursor     int
	offset     int
	selected   map[string]struct{}
	filter     textinput.Model
	filterOn   bool
	selectable bool
	width      int
	height     int
}

func NewList(title, empty string) *ListModel {
	f := textinput.New()
	f.Prompt = "筛选: "
	f.Placeholder = "输入关键字"
	f.CharLimit = 80
	return &ListModel{
		title:      title,
		empty:      empty,
		selected:   make(map[string]struct{}),
		filter:     f,
		selectable: true,
		width:      80,
		height:     20,
	}
}

func (l *ListModel) SetTitle(title string) { l.title = title }
func (l *ListModel) Title() string         { return l.title }
func (l *ListModel) Empty() string         { return l.empty }
func (l *ListModel) SetEmpty(empty string) { l.empty = empty }

func (l *ListModel) SetItems(items []Item) {
	l.items = items
	present := make(map[string]struct{}, len(items))
	for _, item := range items {
		present[item.ID()] = struct{}{}
	}
	for id := range l.selected {
		if _, ok := present[id]; !ok {
			delete(l.selected, id)
		}
	}
	l.recompute()
}

func (l *ListModel) Items() []Item { return l.items }

func (l *ListModel) SetSelectable(selectable bool) {
	l.selectable = selectable
	if !selectable {
		l.ClearSelection()
	}
}

func (l *ListModel) ClearSelection() {
	l.selected = make(map[string]struct{})
}

// ResetFilter clears the filter and leaves filter mode.
func (l *ListModel) ResetFilter() {
	l.filter.SetValue("")
	l.filterOn = false
	l.filter.Blur()
	l.recompute()
}

func (l *ListModel) SetSize(width, height int) {
	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}
	l.width = width
	l.height = height
	l.filter.Width = width - 8
	l.clamp()
}

func (l *ListModel) Filtering() bool { return l.filterOn }

func (l *ListModel) FilterValue() string { return l.filter.Value() }

func (l *ListModel) CurrentItem() Item {
	if len(l.visible) == 0 || l.cursor < 0 || l.cursor >= len(l.visible) {
		return nil
	}
	return l.items[l.visible[l.cursor]]
}

func (l *ListModel) SelectedItems() []Item {
	out := make([]Item, 0, len(l.selected))
	for _, item := range l.items {
		if _, ok := l.selected[item.ID()]; ok {
			out = append(out, item)
		}
	}
	return out
}

func (l *ListModel) SelectedCount() int { return len(l.selected) }
func (l *ListModel) TotalCount() int    { return len(l.items) }
func (l *ListModel) VisibleCount() int  { return len(l.visible) }

func (l *ListModel) IsSelected(item Item) bool {
	if item == nil {
		return false
	}
	_, ok := l.selected[item.ID()]
	return ok
}

func (l *ListModel) recompute() {
	q := strings.ToLower(strings.TrimSpace(l.filter.Value()))
	l.visible = l.visible[:0]
	for i, item := range l.items {
		if q == "" || strings.Contains(strings.ToLower(item.FilterValue()), q) {
			l.visible = append(l.visible, i)
		}
	}
	l.clamp()
}

func (l *ListModel) clamp() {
	if len(l.visible) == 0 {
		l.cursor = 0
		l.offset = 0
		return
	}
	if l.cursor >= len(l.visible) {
		l.cursor = len(l.visible) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
	maxRows := l.maxRows()
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+maxRows {
		l.offset = l.cursor - maxRows + 1
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

func (l *ListModel) maxRows() int {
	h := l.height
	if l.filterOn {
		h--
	}
	rows := h / 2
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (l *ListModel) move(delta int) {
	if len(l.visible) == 0 {
		return
	}
	l.cursor += delta
	l.clamp()
}

func (l *ListModel) toggleCurrent() {
	item := l.CurrentItem()
	if item == nil {
		return
	}
	id := item.ID()
	if _, ok := l.selected[id]; ok {
		delete(l.selected, id)
	} else {
		l.selected[id] = struct{}{}
	}
}

func (l *ListModel) toggleVisible() {
	if len(l.visible) == 0 {
		return
	}
	all := true
	for _, idx := range l.visible {
		if _, ok := l.selected[l.items[idx].ID()]; !ok {
			all = false
			break
		}
	}
	for _, idx := range l.visible {
		id := l.items[idx].ID()
		if all {
			delete(l.selected, id)
		} else {
			l.selected[id] = struct{}{}
		}
	}
}

func (l *ListModel) selectAll() {
	for _, item := range l.items {
		l.selected[item.ID()] = struct{}{}
	}
}

// HandleKey processes list navigation and selection keys. The returned action
// tells the parent app when to open a detail, action menu or previous screen.
func (l *ListModel) HandleKey(msg tea.KeyMsg) (listAction, tea.Cmd) {
	key := msg.String()

	if l.filterOn {
		switch key {
		case "esc":
			l.filter.SetValue("")
			l.filterOn = false
			l.filter.Blur()
			l.recompute()
			return actNone, nil
		case "enter":
			l.filterOn = false
			l.filter.Blur()
			l.recompute()
			return actNone, nil
		case "up", "down":
			l.filterOn = false
			l.filter.Blur()
			l.recompute()
			l.move(map[string]int{"up": -1, "down": 1}[key])
			return actNone, nil
		}
		var cmd tea.Cmd
		l.filter, cmd = l.filter.Update(msg)
		l.recompute()
		return actNone, cmd
	}

	switch key {
	case "/":
		l.filterOn = true
		l.filter.Focus()
		return actNone, textinput.Blink
	case "up", "k":
		l.move(-1)
	case "down", "j":
		l.move(1)
	case "pgup":
		l.move(-l.maxRows())
	case "pgdown":
		l.move(l.maxRows())
	case "home":
		l.cursor = 0
		l.clamp()
	case "end":
		l.cursor = len(l.visible) - 1
		l.clamp()
	case " ":
		if l.selectable {
			l.toggleCurrent()
		}
	case "a":
		if l.selectable {
			l.toggleVisible()
		}
	case "A":
		if l.selectable {
			l.selectAll()
		}
	case "c":
		if l.selectable {
			l.ClearSelection()
		}
	case "enter":
		return actEnter, nil
	case "i":
		return actInfo, nil
	case "o":
		return actOpen, nil
	case "r":
		return actRefresh, nil
	case "esc", "q":
		return actBack, nil
	}
	return actNone, nil
}

// Update feeds cursor-blink and paste messages to the filter input.
func (l *ListModel) Update(msg tea.Msg) tea.Cmd {
	if !l.filterOn {
		return nil
	}
	var cmd tea.Cmd
	l.filter, cmd = l.filter.Update(msg)
	return cmd
}

func (l *ListModel) View(theme Theme) string {
	var b strings.Builder

	if l.filterOn {
		b.WriteString(theme.Input.Render(l.filter.View()))
		b.WriteString("\n")
	}

	if len(l.visible) == 0 {
		msg := l.empty
		if msg == "" {
			msg = "没有内容"
		}
		if l.filter.Value() != "" {
			msg = fmt.Sprintf("没有匹配 “%s” 的结果", l.filter.Value())
		}
		return lipgloss.NewStyle().
			Width(l.width).
			Height(max(1, l.height-1)).
			Align(lipgloss.Center, lipgloss.Center).
			Render(theme.Subtitle.Render(msg))
	}

	start := l.offset
	end := start + l.maxRows()
	if end > len(l.visible) {
		end = len(l.visible)
	}
	lines := 0
	for i := start; i < end; i++ {
		item := l.items[l.visible[i]]
		cursor := "  "
		if i == l.cursor {
			cursor = "▸ "
		}
		check := "  "
		if l.selectable && l.IsSelected(item) {
			check = "✓ "
		}
		title := item.Title()
		if i == l.cursor {
			title = theme.Cursor.Render(title)
		} else if l.IsSelected(item) {
			title = theme.Selected.Render(title)
		} else {
			title = lipgloss.NewStyle().Render(title)
		}
		line1 := fmt.Sprintf("%s%s%s", cursor, check, title)
		if l.width > 0 {
			line1 = ansi.Truncate(line1, l.width, "…")
		}
		b.WriteString(line1)
		b.WriteString("\n")

		desc := item.Description()
		if desc != "" {
			line2 := "    " + theme.Subtitle.Render(desc)
			if l.width > 0 {
				line2 = ansi.Truncate(line2, l.width, "…")
			}
			b.WriteString(line2)
			b.WriteString("\n")
			lines += 2
		} else {
			lines++
		}
		if lines >= l.height-1 {
			break
		}
	}

	if len(l.visible) > l.maxRows() {
		b.WriteString(theme.Help.Render(fmt.Sprintf(
			"  %d-%d / %d", start+1, end, len(l.visible),
		)))
	}
	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
