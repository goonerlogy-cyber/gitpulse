package ui

import "github.com/charmbracelet/lipgloss"

var (
	accent    = lipgloss.Color("#6fe89e")
	subtle    = lipgloss.Color("#5c5c70")
	fg        = lipgloss.Color("#e0e0f0")
	bg        = lipgloss.Color("#0d0d15")
	dimAccent = lipgloss.Color("#3f8f5c")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0d0d15")).
			Background(accent).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().Foreground(subtle)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0d0d15")).
			Background(accent).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(subtle).
				Padding(0, 2)

	sectionTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(accent)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimAccent).
			Padding(1, 2)

	statLabelStyle = lipgloss.NewStyle().Foreground(subtle)
	statValueStyle = lipgloss.NewStyle().Bold(true).Foreground(fg)

	helpStyle = lipgloss.NewStyle().Foreground(subtle)
)
