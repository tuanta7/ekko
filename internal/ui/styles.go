package ui

import "github.com/charmbracelet/lipgloss"

var (
	accentBlue   = lipgloss.Color("#A7E9FF") // pastel cyan
	accentPurple = lipgloss.Color("#D8B7FF") // lavender
	accentGreen  = lipgloss.Color("#BFFFC5") // mint
	textPrimary  = lipgloss.Color("#F7F7FF") // very light off-white
	textMuted    = lipgloss.Color("#9AA2B2") // muted gray-blue

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentBlue).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accentPurple).
			Bold(true).
			PaddingLeft(2)

	normalStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(accentGreen).
			MarginBottom(1)

	transcriptBoxStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				BorderForeground(accentBlue).
				Border(lipgloss.RoundedBorder(), true)

	transcriptTextStyle = lipgloss.NewStyle().
				Foreground(textPrimary)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
)
