package console

import "github.com/charmbracelet/lipgloss"

var (
	bgBase      = lipgloss.Color("#0F172A")
	panelBorder = lipgloss.Color("#334155")
	accentBlue  = lipgloss.Color("#7DD3FC")
	accentMint  = lipgloss.Color("#86EFAC")
	accentRose  = lipgloss.Color("#FCA5A5")
	textPrimary = lipgloss.Color("#F8FAFC")
	textMuted   = lipgloss.Color("#94A3B8")
)

var (
	appFrameStyle = lipgloss.NewStyle().
			Background(bgBase).
			Padding(1, 2)
	headerStyle = lipgloss.NewStyle().
			Foreground(accentBlue).
			Bold(true)
	subheaderStyle = lipgloss.NewStyle().
			Foreground(textMuted)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(panelBorder).
			Padding(0, 1)
	menuItemStyle = lipgloss.NewStyle().
			Foreground(textMuted)
	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(textPrimary).
				Background(lipgloss.Color("#1E293B")).
				Bold(true).
				Padding(0, 1)
	valueStyle = lipgloss.NewStyle().
			Foreground(accentBlue)
	errorStyle = lipgloss.NewStyle().
			Foreground(accentRose).
			Bold(true)
	statusReadyStyle = lipgloss.NewStyle().
				Foreground(textMuted)
	statusRecordingStyle = lipgloss.NewStyle().
				Foreground(accentMint).
				Bold(true)
	transcriptPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(panelBorder).
				Padding(0, 1)
	transcriptStyle = lipgloss.NewStyle().
			Foreground(textPrimary)
	kbdStyle = lipgloss.NewStyle().
			Foreground(textPrimary).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1)
	helpStyle = lipgloss.NewStyle().
			Foreground(textMuted)
)
