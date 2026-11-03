package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func testItems() []Item {
	return []Item{
		MenuItem{Key: "git", Label: "git", Desc: "main"},
		MenuItem{Key: "aria2", Label: "aria2", Desc: "extras"},
		MenuItem{Key: "curl", Label: "curl", Desc: "main"},
	}
}

func TestListSelectionAndFilter(t *testing.T) {
	l := NewList("测试", "空")
	l.SetSize(80, 20)
	l.SetItems(testItems())

	if _, cmd := l.HandleKey(keyRune('/')); cmd == nil {
		// textinput.Blink may be nil in tests; only the state matters.
	}
	if !l.Filtering() {
		t.Fatal("expected filtering mode")
	}
	for _, r := range "git" {
		l.HandleKey(keyRune(r))
	}
	if l.VisibleCount() != 1 {
		t.Fatalf("expected 1 visible item, got %d", l.VisibleCount())
	}
	l.HandleKey(tea.KeyMsg{Type: tea.KeyEnter}) // apply filter
	if l.Filtering() {
		t.Fatal("expected filter mode to end")
	}
	l.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	if got := len(l.SelectedItems()); got != 1 {
		t.Fatalf("expected 1 selected item, got %d", got)
	}
	if got := l.SelectedItems()[0].ID(); got != "git" {
		t.Fatalf("selected %q", got)
	}
}

func TestSelectVisibleOnly(t *testing.T) {
	l := NewList("测试", "空")
	l.SetSize(80, 20)
	l.SetItems(testItems())
	// Filter to names containing "i" (git, aria2).
	l.HandleKey(keyRune('/'))
	for _, r := range "i" {
		l.HandleKey(keyRune(r))
	}
	l.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
	l.HandleKey(keyRune('a'))
	if got := len(l.SelectedItems()); got != 2 {
		t.Fatalf("expected 2 visible selected items, got %d", got)
	}
	l.HandleKey(keyRune('A'))
	if got := len(l.SelectedItems()); got != 3 {
		t.Fatalf("expected all 3 selected items, got %d", got)
	}
	l.HandleKey(keyRune('c'))
	if got := len(l.SelectedItems()); got != 0 {
		t.Fatalf("expected cleared selection, got %d", got)
	}
}

func TestListViewDoesNotPanicWhenEmpty(t *testing.T) {
	l := NewList("空列表", "没有内容")
	l.SetSize(40, 10)
	if got := l.View(DefaultTheme()); got == "" {
		t.Fatal("expected empty-state view")
	}
}

func TestListViewShowsRows(t *testing.T) {
	l := NewList("测试", "空")
	l.SetSize(80, 20)
	l.SetItems(testItems())
	view := l.View(DefaultTheme())
	for _, name := range []string{"git", "aria2", "curl"} {
		if !strings.Contains(view, name) {
			t.Fatalf("expected %q in view:\n%s", name, view)
		}
	}
}
