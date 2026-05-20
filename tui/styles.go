package tui

import "github.com/charmbracelet/lipgloss"

var (
	activeHeaderStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	inactiveHeaderStyle = lipgloss.NewStyle().Faint(true)
	selectedItemStyle   = lipgloss.NewStyle().Background(lipgloss.Color("240"))
	footerStyle         = lipgloss.NewStyle().Faint(true)
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)
