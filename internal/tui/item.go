package tui

// Item is the common interface used by the custom list.
type Item interface {
	ID() string
	Title() string
	Description() string
	FilterValue() string
}

type MenuItem struct {
	Key   string
	Label string
	Desc  string
}

func (m MenuItem) ID() string          { return m.Key }
func (m MenuItem) Title() string       { return m.Label }
func (m MenuItem) Description() string { return m.Desc }
func (m MenuItem) FilterValue() string { return m.Label }
