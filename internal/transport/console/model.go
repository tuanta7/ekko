package console

import (
	"context"
	"errors"
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

	ctx            context.Context
	cancel         context.CancelFunc
	handler        *handler.Handler
	transcriptChan chan string

	// states
	sourcesLoaded       bool
	selectedSourceIndex int
	audioSources        []string
	chunkDuration       time.Duration
	transcriptContent   string
	isRecording         bool
}

func NewModel(handler *handler.Handler) *Model {
	spinnerView := spinner.New()
	spinnerView.Spinner = spinner.Dot

	transcriptView := viewport.New(80, 5)
	transcriptView.SetContent("")

	return &Model{
		screen:        screenMenu,
		spinner:       spinnerView,
		transcript:    transcriptView,
		menuOptions:   menuOptions,
		handler:       handler,
		chunkDuration: defaultChunkDuration,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		sources, err := m.handler.ListSources(context.Background())
		return sourcesLoadedMsg{
			sources: sources,
			err:     err,
		}
	}
}

func (m *Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case optionStartSession:
		if len(m.audioSources) == 0 {
			return m, nil
		}
		return m.startSession()
	case optionAudioSource:
		return m, nil
	case optionChunkDuration:
		return m, nil
	default:
		return m, tea.Quit
	}
}

func (m *Model) startSession() (tea.Model, tea.Cmd) {
	m.screen = screenRecording
	m.isRecording = true
	m.transcriptContent = ""
	m.transcript.SetContent("Recording started...\n")

	source := m.audioSources[m.selectedSourceIndex]
	chunkDuration := m.chunkDuration

	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel
	m.transcriptChan = make(chan string, 100)

	go m.handler.CollectResults(ctx, func(text string) {
		select {
		case m.transcriptChan <- text:
		case <-ctx.Done():
		}
	})

	return m, tea.Batch(
		m.spinner.Tick,
		m.startRecording(source, chunkDuration),
		m.startTranscribe(),
		m.waitForTranscript(),
	)
}

func (m *Model) startRecording(source string, chunkDuration time.Duration) tea.Cmd {
	return func() tea.Msg {
		err := m.handler.StartRecord(m.ctx, chunkDuration, source)
		if err != nil && !errors.Is(err, context.Canceled) {
			return recordingErrorMsg{err: err}
		}
		return nil
	}
}

func (m *Model) startTranscribe() tea.Cmd {
	return func() tea.Msg {
		m.handler.StartTranscribe(m.ctx)
		return nil
	}
}

func (m *Model) waitForTranscript() tea.Cmd {
	return func() tea.Msg {
		select {
		case text, ok := <-m.transcriptChan:
			if !ok {
				return sessionEndMsg{err: nil}
			}
			return transcriptUpdateMsg{text: text}
		case <-m.ctx.Done():
			return sessionEndMsg{err: m.ctx.Err()}
		}
	}
}

func (m *Model) stopSession() (tea.Model, tea.Cmd) {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil

	}

	m.ctx = nil
	m.isRecording = false

	m.handler.Close()

	m.screen = screenMenu
	return m, nil
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
		case "ctrl+c":
			m.stopSession()
			return m, tea.Quit
		case "q", "s":
			return m.stopSession()
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
	case transcriptUpdateMsg:
		if m.isRecording {
			m.transcriptContent += mt.text
			m.transcript.SetContent(m.transcriptContent)
			m.transcript.GotoBottom()
			return m, m.waitForTranscript()
		}
		return m, nil
	case recordingErrorMsg:
		m.transcriptContent += fmt.Sprintf("\n[Error: %v]\n", mt.err)
		m.transcript.SetContent(m.transcriptContent)
		return m.stopSession()
	case sessionEndMsg:
		m.isRecording = false
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
