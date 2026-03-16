package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SpinnerMsg struct {
	Result any
	Err    error
}

type SpinnerModel struct {
	spinner  spinner.Model
	text     string
	Quitting bool
	Result   any
	Err      error
	Action   func() (any, error)
}

func (m SpinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			res, err := m.Action()
			return SpinnerMsg{Result: res, Err: err}
		},
	)
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.Quitting = true
			m.Err = fmt.Errorf("user quit")
			return m, tea.Quit
		}
	case SpinnerMsg:
		m.Result = msg.Result
		m.Err = msg.Err
		m.Quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.Quitting {
		return ""
	}
	str := fmt.Sprintf("\n   %s %s\n", m.spinner.View(), lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.text))
	return str
}

func RunWithSpinner(text string, action func() (any, error)) (any, error) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := SpinnerModel{
		spinner: s,
		text:    text,
		Action:  action,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := finalModel.(SpinnerModel)
	return fm.Result, fm.Err
}

// Success and Fail styling function for lipgloss
func PrintSuccess(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✔ " + msg))
}

func PrintError(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✖ " + msg))
}
func PrintWarning(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⚠ " + msg))
}
func PrintInfo(msg string) {
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("ℹ " + msg))
}
