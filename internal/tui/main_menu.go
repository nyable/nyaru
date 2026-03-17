package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205")).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	menuSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyleMenu     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0, 0, 2)
)

type MenuOption struct {
	TitleStr string
	DescStr  string
}

func (i MenuOption) Title() string       { return i.TitleStr }
func (i MenuOption) Description() string { return i.DescStr }
func (i MenuOption) FilterValue() string { return i.TitleStr }

type MainMenuModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func NewMainMenuModel() MainMenuModel {
	items := []list.Item{
		MenuOption{TitleStr: "Search", DescStr: "搜索可安装的应用 (search)"},
		MenuOption{TitleStr: "List", DescStr: "列出已安装的应用 (list)"},
		MenuOption{TitleStr: "Status", DescStr: "检查可更新的应用 (status)"},
		MenuOption{TitleStr: "Update All", DescStr: "更新所有应用 (update)"},
		MenuOption{TitleStr: "Buckets", DescStr: "存储桶管理 (bucket)"},
		MenuOption{TitleStr: "Cache", DescStr: "缓存清理 (cache)"},
		MenuOption{TitleStr: "Exit", DescStr: "退出程序 (exit)"},
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = menuSelectedItemStyle
	d.Styles.SelectedDesc = menuSelectedItemStyle.Copy().Foreground(lipgloss.Color("252"))

	m := list.New(items, d, 0, 0)
	m.Title = "Nyaru - Scoop TUI Main Menu"
	m.Styles.Title = titleStyle
	m.SetShowStatusBar(false)
	m.SetFilteringEnabled(false)
	m.SetShowHelp(false)

	return MainMenuModel{
		list: m,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(MenuOption)
			if ok {
				m.choice = i.TitleStr
				m.quitting = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainMenuModel) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View() + "\n" + helpStyleMenu.Render("enter: select • q: quit"))
}

func RunMainMenu() (string, error) {
	m := NewMainMenuModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	if fm, ok := finalModel.(MainMenuModel); ok {
		if fm.choice == "Exit" {
			return "", nil
		}
		return fm.choice, nil
	}
	return "", nil
}
