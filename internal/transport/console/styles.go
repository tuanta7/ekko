package console

import "github.com/charmbracelet/lipgloss"

var (
	accentCyan  = lipgloss.Color("#A7E9FF")
	accentMint  = lipgloss.Color("#BFFFC5")
	accentPink  = lipgloss.Color("#FFB7D5")
	textPrimary = lipgloss.Color("#F7F7FF")
	textMuted   = lipgloss.Color("#9AA2B2")
	textError   = lipgloss.Color("#FF6B6B")
)

var (
	logoStyle       = lipgloss.NewStyle().Foreground(accentCyan).Bold(true)
	selectedStyle   = lipgloss.NewStyle().Foreground(accentMint)
	normalStyle     = lipgloss.NewStyle().Foreground(textMuted)
	valueStyle      = lipgloss.NewStyle().Foreground(accentCyan)
	statusStyle     = lipgloss.NewStyle().Foreground(accentPink)
	helpStyle       = lipgloss.NewStyle().Foreground(textMuted)
	errorStyle      = lipgloss.NewStyle().Foreground(textError)
	transcriptStyle = lipgloss.NewStyle().Foreground(textPrimary)
)

const logo = `
     _    _        
  __| | _| | _____ 
 / _  |/ /| |/ / _ \
|  __/|  <|   < (_) |
 \___||_\_\_|\_\___/
`
