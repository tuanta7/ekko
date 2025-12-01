package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/tuanta7/transcript/internal/core"
)

type screen int

const (
	screenMenu screen = iota
	screenRecording
)

type Model struct {
	screen        screen
	cursor        int
	menuOptions   []string
	chunkDuration time.Duration
	errorMsg      string

	spinner           spinner.Model
	transcript        viewport.Model
	transcriptContent string

	app    *core.Application
	stream <-chan string
}

func NewModel(app *core.Application) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	vp := viewport.New(90, 5)
	vp.SetContent("")

	return &Model{
		screen:        screenMenu,
		menuOptions:   []string{"Start Session", "Chunk Duration", "Exit"},
		spinner:       sp,
		transcript:    vp,
		app:           app,
		chunkDuration: 10 * time.Second,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0:
		m.screen = screenRecording
		m.transcriptContent = ""
		m.transcript.SetContent("")
		m.transcript.YOffset = 0
		m.errorMsg = ""

		var err error
		m.stream, err = m.app.Start(m.chunkDuration)
		if err != nil {
			return m, tea.Quit
		}

		return m, tea.Batch(m.spinner.Tick, m.waitForTranscript())
	case 1:
		// chunk duration, no action on selection
		return m, nil
	case len(m.menuOptions) - 1:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *Model) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenMenu:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.menuOptions)-1 {
				m.cursor++
			}
		case "left":
			if m.cursor == 1 {
				if m.chunkDuration > time.Second {
					m.chunkDuration -= time.Second
				}
			}
		case "right":
			if m.cursor == 1 {
				if m.chunkDuration < 60*time.Second {
					m.chunkDuration += time.Second
				}
			}
		case "enter":
			return m.handleMenuSelection()
		}
	case screenRecording:
		switch msg.String() {
		case "ctrl+c", "q":
			_, _ = m.app.Stop()
			return m, tea.Quit
		case "s", "S":
			return m, m.sessionEnd()
		default:
			var cmd tea.Cmd
			m.transcript, cmd = m.transcript.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mt := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyEvent(mt)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(mt)
		return m, cmd
	case transcriptChunkMsg:
		m.transcriptContent += mt.Text + "\n"
		wrapped := wordwrap.String(m.transcriptContent, m.transcript.Width-3)
		m.transcript.SetContent(wrapped)
		m.transcript.GotoBottom()
		return m, m.waitForTranscript()
	case sessionEndMsg:
		m.screen = screenMenu
		if mt.Error != nil {
			m.errorMsg = fmt.Sprintf("Error: %v", mt.Error)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	var b strings.Builder

	switch m.screen {
	case screenMenu:
		b.WriteString(titleStyle.Render("Audio Transcription"))
		b.WriteString("\n")

		for i, choice := range m.menuOptions {
			cursor := " "
			label := choice
			if choice == "Chunk Duration" {
				label = fmt.Sprintf("Chunk Duration: %ds", int(m.chunkDuration.Seconds()))
			}
			if m.cursor == i {
				cursor = ">"
				label = selectedStyle.Render(label)
			} else {
				label = normalStyle.Render(label)
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, label))
		}

		if m.errorMsg != "" {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render(m.errorMsg))
			b.WriteString("\n")
		}

		b.WriteString(helpStyle.Render("up/down: navigate • left/right: adjust duration • enter: select • q: quit"))
	case screenRecording:
		b.WriteString(titleStyle.Render("Recording Session"))
		b.WriteString("\n")
		b.WriteString(statusStyle.Render(m.spinner.View() + " Recording..."))
		b.WriteString("\n")
		b.WriteString(transcriptBoxStyle.Render(transcriptTextStyle.Render(m.transcript.View())))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓: scroll • s: stop and save • q: quit"))
	}

	return b.String()
}
