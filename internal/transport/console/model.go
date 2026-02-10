package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
	"github.com/tuanta7/ekko/internal/manager"
	"github.com/tuanta7/ekko/pkg/logger"
)

type screen int

const (
	screenMenu screen = iota
	screenRecording
)

type Model struct {
	screen        screen
	cursor        int
	spinner       spinner.Model
	transcript    viewport.Model
	menuOptions   []string
	chunkDuration time.Duration

	sessionStart      time.Time
	transcriptContent string
	isStopping        bool
	errorMsg          string

	handler *manager.Handler
	stream  <-chan string
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *logger.Logger
}

func NewModel(app *manager.Handler, logger *logger.Logger) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	vp := viewport.New(80, 5)
	vp.SetContent("")

	return &Model{
		screen:        screenMenu,
		menuOptions:   []string{"Start Session", "Chunk Duration", "Exit"},
		spinner:       sp,
		transcript:    vp,
		handler:       app,
		chunkDuration: 10 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
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
		m.sessionStart = time.Now()

		var err error
		m.stream, err = m.handler.Start(m.ctx, m.chunkDuration)
		if err != nil {
			return m, m.sessionEnd(err)
		}

		return m, tea.Batch(m.spinner.Tick, m.waitForTranscript())
	case 1:
		return m, nil
	default:
		return m, tea.Quit
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
			m.cancel()
			if err := m.handler.Stop(); err != nil {
				m.logger.Error(fmt.Sprintf("Failed to stop session: %v", err))
			}
			return m, tea.Quit
		case "s", "S":
			m.cancel()
			m.isStopping = true
			err := m.handler.Stop()
			m.isStopping = false
			return m, m.sessionEnd(err)
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
	case transcriptMsg:
		_ = m.handler.SaveChunk(mt.chunk)

		m.transcriptContent += mt.chunk
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
		b.WriteString(logoStyle.Render(logo))
		b.WriteString("\n")

		for i, choice := range m.menuOptions {
			label := choice
			if choice == "Chunk Duration" {
				label = fmt.Sprintf("Chunk Duration: %s", valueStyle.Render(fmt.Sprintf("%ds", int(m.chunkDuration.Seconds()))))
			}

			if m.cursor == i {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("> %s", label)) + "\n")
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %s", label)) + "\n")
			}
		}

		if m.errorMsg != "" {
			b.WriteString("\n" + errorStyle.Render(m.errorMsg) + "\n")
		}

		b.WriteString("\n" + helpStyle.Render("↑/↓ navigate  ←/→ adjust  enter select  q quit"))

	case screenRecording:
		elapsed := time.Since(m.sessionStart).Round(time.Second)
		status := fmt.Sprintf("%s %s  %s",
			m.spinner.View(),
			statusStyle.Render("Recording"),
			valueStyle.Render(elapsed.String()))
		b.WriteString(status + "\n\n")

		if m.isStopping {
			b.WriteString(statusStyle.Render("Saving transcript...") + "\n\n")
		}

		b.WriteString(transcriptStyle.Render(m.transcript.View()))
		b.WriteString("\n\n" + helpStyle.Render("↑/↓ scroll  s stop  q quit"))
	}

	return b.String()
}
