package console

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tuanta7/ekko/internal/handler"
)

type Model struct {
	screen      screen
	cursor      int
	spinner     spinner.Model
	transcript  viewport.Model
	menuOptions []string

	ctx     context.Context
	cancel  context.CancelFunc
	handler *handler.Handler

	// states
	sourcesLoaded       bool
	selectedSourceIndex int
	audioSources        []string
	chunkDuration       time.Duration
}

func NewModel(handler *handler.Handler) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	spinnerView := spinner.New()
	spinnerView.Spinner = spinner.Dot

	transcriptView := viewport.New(80, 5)
	transcriptView.SetContent("")

	return &Model{
		screen:        screenMenu,
		spinner:       spinnerView,
		transcript:    transcriptView,
		menuOptions:   menuOptions,
		ctx:           ctx,
		cancel:        cancel,
		handler:       handler,
		chunkDuration: defaultChunkDuration,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		sources, err := m.handler.ListSources(m.ctx)
		return sourcesLoadedMsg{
			sources: sources,
			err:     err,
		}
	}
}

func (m *Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case optionStartSession:
		return m, nil
	case optionAudioSource:
		return m, nil
	case optionChunkDuration:
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
			if m.cursor < optionExit {
				m.cursor++
			}
		case "left":
			if m.cursor == optionAudioSource && len(m.audioSources) > 0 {
				if m.selectedSourceIndex > 0 {
					m.selectedSourceIndex--
				}
			} else if m.cursor == optionChunkDuration {
				if m.chunkDuration > time.Second {
					m.chunkDuration -= time.Second
				}
			}
		case "right":
			if m.cursor == optionAudioSource && len(m.audioSources) > 0 {
				if m.selectedSourceIndex < len(m.audioSources)-1 {
					m.selectedSourceIndex++
				}
			} else if m.cursor == optionChunkDuration {
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
			return m, tea.Quit
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
	case sourcesLoadedMsg:
		if mt.err == nil {
			m.sourcesLoaded = true
			m.audioSources = mt.sources
			m.selectedSourceIndex = 0
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(mt)
		return m, cmd
	case sessionEndMsg:
		m.screen = screenMenu
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
			if choice == "Audio Source" {
				if len(m.audioSources) > 0 {
					sourceName := m.audioSources[m.selectedSourceIndex]
					label = fmt.Sprintf("Audio Source: %s", valueStyle.Render(sourceName))
				} else if m.sourcesLoaded {
					label = fmt.Sprintf("Audio Source: %s", valueStyle.Render("No sources found"))
				} else {
					label = fmt.Sprintf("Audio Source: %s", valueStyle.Render("Loading..."))
				}
			} else if choice == "Chunk Duration" {
				label = fmt.Sprintf("Chunk Duration: %s", valueStyle.Render(fmt.Sprintf("%ds", int(m.chunkDuration.Seconds()))))
			}

			if m.cursor == i {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("> %s", label)) + "\n")
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %s", label)) + "\n")
			}
		}

		b.WriteString("\n" + helpStyle.Render("↑/↓ navigate  ←/→ adjust  enter select  q quit"))

	case screenRecording:
		b.WriteString(transcriptStyle.Render(m.transcript.View()))
		b.WriteString("\n\n" + helpStyle.Render("↑/↓ scroll  s stop  q quit"))
	}

	return b.String()
}
