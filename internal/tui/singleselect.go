package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SingleSelectModel struct {
	Options []string
	Cursor  int
	Title   string
	Done    bool
	Quit    bool
}

func NewSingleSelect(title string, options []string) SingleSelectModel {
	return SingleSelectModel{
		Options: options,
		Title:   title,
	}
}

func (m SingleSelectModel) Init() tea.Cmd {
	return nil
}

func (m SingleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, nil
}

func (m SingleSelectModel) View() string {
	if m.Done || m.Quit {
		return ""
	}

	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.Title) + "\n\n")

	start := 0
	end := len(m.Options)

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
		style := lipgloss.NewStyle()
		if m.Cursor == i {
			cursor = focusedStyle.Render(">")
			style = focusedStyle
		}

		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(m.Options[i])))
	}

	b.WriteString(helpStyle.Render("\n↑/↓: 移动空间 • 回车: 确认 • q/ctrl+c: 退出\n"))
	return b.String()
}

func RunSingleSelect(title string, options []string) (string, error) {
	p := tea.NewProgram(NewSingleSelect(title, options))
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	model := m.(SingleSelectModel)
	if model.Quit {
		return "", fmt.Errorf("user quit")
	}
	return model.Options[model.Cursor], nil
}
