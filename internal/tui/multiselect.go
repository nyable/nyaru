package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type MultiSelectModel struct {
	Options  []string
	Cursor   int
	Selected map[int]struct{}
	Title    string
	err      error
	Done     bool
	Quit     bool
}

func NewMultiSelect(title string, options []string) MultiSelectModel {
	return MultiSelectModel{
		Options:  options,
		Selected: make(map[int]struct{}),
		Title:    title,
	}
}

func (m MultiSelectModel) Init() tea.Cmd {
	return nil
}

func (m MultiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quit = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Options)-1 {
				m.Cursor++
			}
		case "enter":
			m.Done = true
			return m, tea.Quit
		case " ":
			_, ok := m.Selected[m.Cursor]
			if ok {
				delete(m.Selected, m.Cursor)
			} else {
				m.Selected[m.Cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m MultiSelectModel) View() string {
	if m.Done || m.Quit {
		return ""
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.Title) + "\n\n")

	start := 0
	end := len(m.Options)

	// Display only a window of items if the list is long
	if len(m.Options) > 20 {
		start = m.Cursor - 10
		if start < 0 {
			start = 0
		}
		end = start + 20
		if end > len(m.Options) {
			end = len(m.Options)
			start = end - 20
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		cursor := " "
		if m.Cursor == i {
			cursor = focusedStyle.Render(">")
		}

		checked := " "
		if _, ok := m.Selected[i]; ok {
			checked = "x"
		}

		style := lipgloss.NewStyle()
		if m.Cursor == i {
			style = focusedStyle
		} else if _, ok := m.Selected[i]; ok {
			style = selectedItemStyle
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, style.Render(m.Options[i])))
	}

	b.WriteString(helpStyle.Render("\n↑/↓: 移动空间 • 空格: 选择 • 回车: 确认 • q/ctrl+c: 退出\n"))
	return b.String()
}

func RunMultiSelect(title string, options []string) ([]string, error) {
	p := tea.NewProgram(NewMultiSelect(title, options))
	m, err := p.Run()
	if err != nil {
		return nil, err
	}
	model := m.(MultiSelectModel)
	if model.Quit {
		return nil, fmt.Errorf("user quit")
	}
	var selected []string
	for i := range model.Options {
		if _, ok := model.Selected[i]; ok {
			selected = append(selected, model.Options[i])
		}
	}
	return selected, nil
}
