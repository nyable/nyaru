package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nyable/nyaru/internal/models"
	"github.com/nyable/nyaru/internal/utils"
)

var (
	docStyle         = lipgloss.NewStyle().Margin(0, 1)
	detailTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).MarginBottom(1)
	detailLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	detailValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

type UIState int

const (
	StateList UIState = iota
	StateDetail
)

type ListModel struct {
	list        list.Model
	state       UIState
	selectedItem list.Item
	choices     []list.Item
	infoFunc    func(string) (string, error)
	infoContent string
	quitting    bool
}



func NewListModel(title string, items []list.Item, infoFunc func(string) (string, error)) ListModel {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetHeight(2) // Default is 2 lines for title+desc
	d.SetSpacing(1) // Keep it readable

	m := list.New(items, d, 0, 0)
	m.Title = title
	// Remove extra padding from the list itself
	m.Styles.Title.MarginLeft(0)
	m.Styles.PaginationStyle.PaddingLeft(0)
	m.Styles.HelpStyle.PaddingLeft(0)
	// Start in filtering mode for better search experience (fzf-like)
	m.FilterInput.Placeholder = "Type to filter apps..."
	m.SetFilteringEnabled(true)

	return ListModel{
		list:     m,
		state:    StateList,
		infoFunc: infoFunc,
	}
}


func (m ListModel) Init() tea.Cmd {
	return nil
}

type infoMsg struct {
	content string
	err     error
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case infoMsg:
		if msg.err != nil {
			m.infoContent = "Error: " + msg.err.Error()
		} else {
			m.infoContent = msg.content
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Handle key presses when not in filtering mode (or when exiting it)
		if m.list.FilterState() != list.Filtering {
			switch m.state {
			case StateList:
				switch msg.String() {
				case "i":
					if m.infoFunc == nil {
						return m, nil
					}
					item := m.list.SelectedItem()
					if item == nil {
						return m, nil
					}
					m.selectedItem = item

					name := ""
					if i, ok := item.(models.AppInfo); ok {
						name = i.Name
					} else if b, ok := item.(models.BucketResult); ok {
						name = b.Name
					} else if c, ok := item.(models.CacheResult); ok {
						name = c.Name
					}

					if name == "" {
						return m, nil
					}

					m.state = StateDetail
					m.infoContent = "Loading..."

					return m, func() tea.Msg {
						content, err := m.infoFunc(name)
						return infoMsg{content: content, err: err}
					}
				case " ":
					idx := m.list.Index()
					if idx >= 0 && idx < len(m.list.Items()) {
						item := m.list.Items()[idx]
						if i, ok := item.(models.AppInfo); ok {
							i.Selected = !i.Selected
							m.list.SetItem(idx, i)
						} else if i, ok := item.(models.CacheResult); ok {
							i.Selected = !i.Selected
							m.list.SetItem(idx, i)
						} else if i, ok := item.(models.BucketResult); ok {
							i.Selected = !i.Selected
							m.list.SetItem(idx, i)
						}
					}
					m.updateTitle()
					return m, nil
				case "a":
					// Toggle select all/none
					items := m.list.Items()
					allSelected := true
					for _, item := range items {
						if i, ok := item.(models.AppInfo); ok && !i.Selected {
							allSelected = false
							break
						} else if i, ok := item.(models.CacheResult); ok && !i.Selected {
							allSelected = false
							break
						} else if i, ok := item.(models.BucketResult); ok && !i.Selected {
							allSelected = false
							break
						}
					}

					for idx, item := range items {
						if i, ok := item.(models.AppInfo); ok {
							i.Selected = !allSelected
							m.list.SetItem(idx, i)
						} else if i, ok := item.(models.CacheResult); ok {
							i.Selected = !allSelected
							m.list.SetItem(idx, i)
						} else if i, ok := item.(models.BucketResult); ok {
							i.Selected = !allSelected
							m.list.SetItem(idx, i)
						}
					}
					m.updateTitle()
					return m, nil
				case "enter":
					m.choices = m.getSelectedItems()
					if len(m.choices) == 0 {
						if i := m.list.SelectedItem(); i != nil {
							m.choices = append(m.choices, i)
						}
					}
					m.quitting = true
					return m, tea.Quit
				}

			case StateDetail:
				switch msg.String() {
				case "esc", "backspace", "q":
					m.state = StateList
					return m, nil
				case "o":
					// Open URL
					url := ""
					if b, ok := m.selectedItem.(models.BucketResult); ok {
						url = b.Source
					} else if m.infoContent != "" {
						// Extract homepage/website from info content
						lines := strings.Split(m.infoContent, "\n")
						for _, line := range lines {
							lowerLine := strings.ToLower(line)
							if strings.Contains(lowerLine, "homepage") || strings.Contains(lowerLine, "website") {
								parts := strings.Split(line, ":")
								if len(parts) > 1 {
									potentialURL := strings.TrimSpace(strings.Join(parts[1:], ":"))
									potentialURL = strings.Trim(potentialURL, "\" ',")
									if idx := strings.Index(potentialURL, " "); idx != -1 {
										potentialURL = potentialURL[:idx]
									}
									url = potentialURL
									break
								}
							}
						}
					}

					if url != "" {
						utils.OpenBrowser(url)
					}
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		// Calculate available height: Total - Margin(v) - Help(1) - Title(1) - Status(1)
		// Actually bubbletea's list.Model handles its own sub-elements if SetSize is correct.
		// However, we can fine-tune it.
		listHeight := msg.Height - v
		m.list.SetSize(msg.Width-h, listHeight)
		detailValueStyle = detailValueStyle.Width(msg.Width - h - 4) // Adjust width for detail view

	}

	m.list, cmd = m.list.Update(msg)
	
	// Also update title on other msgs just in case
	m.updateTitle()

	return m, cmd
}

func (m *ListModel) updateTitle() {
	selectedCount := 0
	var selectedBytes int64
	hasCache := false

	for _, item := range m.list.Items() {
		if i, ok := item.(models.AppInfo); ok && i.Selected {
			selectedCount++
		} else if i, ok := item.(models.CacheResult); ok {
			hasCache = true
			if i.Selected {
				selectedCount++
				selectedBytes += i.Length
			}
		} else if i, ok := item.(models.BucketResult); ok && i.Selected {
			selectedCount++
		}
	}
	
	baseTitle := m.list.Title
	if strings.Contains(baseTitle, " [") {
		baseTitle = baseTitle[:strings.LastIndex(baseTitle, " [")]
	}
	
	if selectedCount > 0 {
		status := fmt.Sprintf("[%d Selected]", selectedCount)
		if hasCache {
			// (Note: utils might be needed for size formatting, but we can do a simple one here or pass it)
			// Actually, better to import utils in list_view.go if possible or use a simple conversion
			// Let's use HumanSizeToBytes's counterpart or just raw MiB for now if we don't want to add more imports
			// Actually I'll just use a simple inline fmt for now to keep it lean.
			status = fmt.Sprintf("[%d Selected, %.2f MiB Total]", selectedCount, float64(selectedBytes)/(1024*1024))
		}
		m.list.Title = fmt.Sprintf("%s %s", baseTitle, status)
	} else {
		m.list.Title = baseTitle
	}
}

func (m ListModel) getSelectedItems() []list.Item {
	var selected []list.Item
	for _, item := range m.list.Items() {
		if i, ok := item.(models.AppInfo); ok && i.Selected {
			selected = append(selected, i)
		} else if i, ok := item.(models.CacheResult); ok && i.Selected {
			selected = append(selected, i)
		} else if i, ok := item.(models.BucketResult); ok && i.Selected {
			selected = append(selected, i)
		}
	}
	return selected
}


func (m ListModel) View() string {
	if m.quitting {
		return ""
	}

	if m.state == StateList {
		return docStyle.Render(m.list.View() + "\n" + renderListHelp())
	}

	if m.state == StateDetail && m.selectedItem != nil {
		title := "Details"
		if i, ok := m.selectedItem.(models.AppInfo); ok {
			title = "App Details: " + i.Name
		} else if b, ok := m.selectedItem.(models.BucketResult); ok {
			title = "Bucket Details: " + b.Name
		} else if c, ok := m.selectedItem.(models.CacheResult); ok {
			title = "Cache Details: " + c.Name
		}

		view := detailTitleStyle.Render(title) + "\n"

		if b, ok := m.selectedItem.(models.BucketResult); ok {
			// Specific view for buckets as requested
			view += renderDetailLine("Name", b.Name) + "\n"
			view += renderDetailLine("Source", b.Source) + "\n"
			view += renderDetailLine("Manifests", fmt.Sprintf("%d", b.Manifests)) + "\n"
			view += renderDetailLine("Updated", b.Updated.DateTime) + "\n"
		} else if m.infoContent != "" {
			// Define the fields we want to display
			fields := []struct {
				Key    string
				Labels []string // Possible keys in JSON or plain text
			}{
				{"Name", []string{"name", "Name"}},
				{"Description", []string{"description", "Description"}},
				{"Version", []string{"version", "Version"}},
				{"Bucket", []string{"bucket", "Source", "Bucket"}},
				{"Website", []string{"website", "Website"}},
				{"License", []string{"license", "License"}},
				{"Binaries", []string{"binaries", "Binaries"}},
				{"Notes", []string{"notes", "Notes"}},
			}

			// Try to parse as JSON first (sfsu)
			var jsonMap map[string]interface{}
			isJSON := json.Unmarshal([]byte(m.infoContent), &jsonMap) == nil

			if isJSON {
				for _, f := range fields {
					val := ""
					for _, label := range f.Labels {
						if v, ok := jsonMap[label]; ok && v != nil {
							val = fmt.Sprintf("%v", v)
							break
						}
					}
					if f.Key == "Notes" {
						if val != "" {
							view += "\n" + detailLabelStyle.Render("Notes:") + "\n" + detailValueStyle.Render(val) + "\n"
						}
					} else {
						view += renderDetailLine(f.Key, val) + "\n"
					}
				}
			} else {
				// Parse plain text (scoop)
				lines := strings.Split(m.infoContent, "\n")
				data := make(map[string]string)
				var currentKey string
				var currentVal strings.Builder

				for _, line := range lines {
					if strings.Contains(line, " : ") {
						// Save previous
						if currentKey != "" {
							data[currentKey] = strings.TrimSpace(currentVal.String())
						}
						parts := strings.SplitN(line, " : ", 2)
						currentKey = strings.TrimSpace(parts[0])
						currentVal.Reset()
						currentVal.WriteString(parts[1])
					} else if currentKey != "" && strings.HasPrefix(line, "              ") {
						currentVal.WriteString("\n")
						currentVal.WriteString(strings.TrimSpace(line))
					}
				}
				if currentKey != "" {
					data[currentKey] = strings.TrimSpace(currentVal.String())
				}

				for _, f := range fields {
					val := ""
					for _, label := range f.Labels {
						if v, ok := data[label]; ok {
							val = v
							break
						}
					}
					if f.Key == "Notes" {
						if val != "" {
							view += "\n" + detailLabelStyle.Render("Notes:") + "\n" + detailValueStyle.Render(val) + "\n"
						}
					} else {
						view += renderDetailLine(f.Key, val) + "\n"
					}
				}
			}
		} else {
			if i, ok := m.selectedItem.(models.AppInfo); ok {
				lines := []string{
					renderDetailLine("Name", i.Name),
					renderDetailLine("Version", i.Version),
					renderDetailLine("Bucket", i.Bucket),
					renderDetailLine("Installed", fmt.Sprintf("%v", i.Installed)),
				}
				for _, l := range lines {
					view += l + "\n"
				}
			} else if b, ok := m.selectedItem.(models.BucketResult); ok {
				lines := []string{
					renderDetailLine("Name", b.Name),
					renderDetailLine("Source", b.Source),
					renderDetailLine("Manifests", fmt.Sprintf("%d", b.Manifests)),
					renderDetailLine("Updated", b.Updated.DateTime),
				}
				for _, l := range lines {
					view += l + "\n"
				}
			} else if c, ok := m.selectedItem.(models.CacheResult); ok {
				lines := []string{
					renderDetailLine("Name", c.Name),
					renderDetailLine("Version", c.Version),
					renderDetailLine("Size", c.FormatSize),
				}
				for _, l := range lines {
					view += l + "\n"
				}
			}
		}


		view += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("o: open url • Esc/Backspace/q: back")

		return docStyle.Render(view)
	}


	return ""
}

func renderListHelp() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Space: select • a: all • Enter: confirm • i: info • /: filter")
}

func renderDetailLine(label, value string) string {
	return detailLabelStyle.Render(fmt.Sprintf("%-12s: ", label)) + detailValueStyle.Render(value)
}

// RunListInteractive shows the list UI and returns the selected items slice.
func RunListInteractive(title string, items []list.Item, infoFunc func(string) (string, error)) ([]list.Item, error) {
	m := NewListModel(title, items, infoFunc)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if fm, ok := finalModel.(ListModel); ok {
		return fm.choices, nil
	}
	return nil, nil // user quit or didn't select
}


