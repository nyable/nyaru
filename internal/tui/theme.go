package tui

import "github.com/charmbracelet/lipgloss"

// Theme centralizes every color and style used by the TUI.
type Theme struct {
	Primary lipgloss.AdaptiveColor
	Success lipgloss.AdaptiveColor
	Warning lipgloss.AdaptiveColor
	Error   lipgloss.AdaptiveColor
	Muted   lipgloss.AdaptiveColor
	Border  lipgloss.AdaptiveColor

	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Label      lipgloss.Style
	Value      lipgloss.Style
	Help       lipgloss.Style
	Status     lipgloss.Style
	ErrorText  lipgloss.Style
	SuccessTxt lipgloss.Style
	WarnText   lipgloss.Style
	Selected   lipgloss.Style
	Cursor     lipgloss.Style
	Modal      lipgloss.Style
	Input      lipgloss.Style
}

func DefaultTheme() Theme {
	primary := lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#5EEAD4"}
	success := lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#4ADE80"}
	warning := lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}
	errColor := lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#F87171"}
	muted := lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	border := lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#374151"}

	return Theme{
		Primary:    primary,
		Success:    success,
		Warning:    warning,
		Error:      errColor,
		Muted:      muted,
		Border:     border,
		Title:      lipgloss.NewStyle().Bold(true).Foreground(primary),
		Subtitle:   lipgloss.NewStyle().Foreground(muted),
		Label:      lipgloss.NewStyle().Bold(true).Foreground(primary),
		Value:      lipgloss.NewStyle(),
		Help:       lipgloss.NewStyle().Foreground(muted),
		Status:     lipgloss.NewStyle().Foreground(muted),
		ErrorText:  lipgloss.NewStyle().Foreground(errColor),
		SuccessTxt: lipgloss.NewStyle().Foreground(success),
		WarnText:   lipgloss.NewStyle().Foreground(warning),
		Selected:   lipgloss.NewStyle().Bold(true).Foreground(primary),
		Cursor:     lipgloss.NewStyle().Bold(true).Foreground(primary),
		Modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(1, 2),
		Input: lipgloss.NewStyle().Foreground(primary),
	}
}
