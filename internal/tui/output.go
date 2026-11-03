package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#4ADE80"})
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B91C1C", Dark: "#F87171"})
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"})
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0F766E", Dark: "#5EEAD4"})
)

func PrintSuccess(msg string) { fmt.Println(successStyle.Render("✓ " + msg)) }
func PrintError(msg string)   { fmt.Println(errorStyle.Render("✗ " + msg)) }
func PrintWarning(msg string) { fmt.Println(warningStyle.Render("⚠ " + msg)) }
func PrintInfo(msg string)    { fmt.Println(infoStyle.Render("ℹ " + msg)) }
