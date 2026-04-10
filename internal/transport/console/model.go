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
	"github.com/charmbracelet/lipgloss"
	"github.com/tuanta7/ekko/internal/handler"
)

type Model struct {
	screen      screen
	cursor      int
	spinner     spinner.Model
	transcript  viewport.Model
	menuOptions []string
	width       int
	height      int

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
	lastError           string
}

func NewModel(handler *handler.Handler) *Model {
	spinnerView := spinner.New()
	spinnerView.Spinner = spinner.Dot

	transcriptView := viewport.New(60, 10)
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
	m.lastError = ""
	m.transcriptContent = ""
	m.transcript.SetContent("Listening for speech...\n")

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
		m.startRecording(ctx, source, chunkDuration),
		m.startTranscribe(ctx),
		m.waitForTranscript(ctx, m.transcriptChan),
	)
}

func (m *Model) startRecording(ctx context.Context, source string, chunkDuration time.Duration) tea.Cmd {
	return func() tea.Msg {
		err := m.handler.StartRecord(ctx, chunkDuration, source)
		if err != nil && !errors.Is(err, context.Canceled) {
			return recordingErrorMsg{err: err}
		}
		return nil
	}
}

func (m *Model) startTranscribe(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		m.handler.StartTranscribe(ctx)
		return nil
	}
}

func (m *Model) waitForTranscript(ctx context.Context, transcriptChan <-chan string) tea.Cmd {
	return func() tea.Msg {
		if ctx == nil {
			return sessionEndMsg{err: context.Canceled}
		}
		select {
		case text, ok := <-transcriptChan:
			if !ok {
				return sessionEndMsg{err: nil}
			}
			return transcriptUpdateMsg{text: text}
		case <-ctx.Done():
			return sessionEndMsg{err: ctx.Err()}
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
	m.transcriptChan = nil
	m.screen = screenMenu
	return m, nil
}

func (m *Model) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenMenu:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "ctrl+n":
			if m.cursor < optionExit {
				m.cursor++
			}
		case "left", "h":
			if m.cursor == optionAudioSource && len(m.audioSources) > 0 {
				if m.selectedSourceIndex > 0 {
					m.selectedSourceIndex--
				}
			} else if m.cursor == optionChunkDuration {
				if m.chunkDuration > time.Second {
					m.chunkDuration -= time.Second
				}
			}
		case "right", "l":
			if m.cursor == optionAudioSource && len(m.audioSources) > 0 {
				if m.selectedSourceIndex < len(m.audioSources)-1 {
					m.selectedSourceIndex++
				}
			} else if m.cursor == optionChunkDuration {
				if m.chunkDuration < 60*time.Second {
					m.chunkDuration += time.Second
				}
			}
		case "enter", " ":
			return m.handleMenuSelection()
		}
	case screenRecording:
		switch msg.String() {
		case "ctrl+c":
			m.stopSession()
			return m, tea.Quit
		case "q", "s", "esc":
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
	case tea.WindowSizeMsg:
		m.width = mt.Width
		m.height = mt.Height
		m.resizeTranscript()
		return m, nil
	case sourcesLoadedMsg:
		m.sourcesLoaded = true
		if mt.err == nil {
			m.audioSources = mt.sources
			m.selectedSourceIndex = 0
			m.lastError = ""
		} else {
			m.lastError = fmt.Sprintf("Failed to load audio sources: %v", mt.err)
		}
		return m, nil
	case spinner.TickMsg:
		if !m.isRecording {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(mt)
		return m, cmd
	case transcriptUpdateMsg:
		if m.isRecording {
			text := strings.TrimSpace(mt.text)
			if text == "" {
				return m, m.waitForTranscript(m.ctx, m.transcriptChan)
			}
			m.transcriptContent += fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), text)
			m.transcript.SetContent(m.transcriptContent)
			m.transcript.GotoBottom()
			return m, m.waitForTranscript(m.ctx, m.transcriptChan)
		}
		return m, nil
	case recordingErrorMsg:
		m.lastError = fmt.Sprintf("Recording failed: %v", mt.err)
		return m.stopSession()
	case sessionEndMsg:
		if mt.err != nil && !errors.Is(mt.err, context.Canceled) {
			m.lastError = fmt.Sprintf("Session ended with error: %v", mt.err)
		}
		m.isRecording = false
		m.cancel = nil
		m.ctx = nil
		m.transcriptChan = nil
		m.screen = screenMenu
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	switch m.screen {
	case screenMenu:
		return m.renderMenu()
	case screenRecording:
		return m.renderRecording()
	}

	return ""
}

func (m *Model) renderMenu() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("Ekko Live"),
		subheaderStyle.Render("Real-time transcription in your terminal"),
		"",
		statusReadyStyle.Render("Status: Ready"),
	)

	if m.lastError != "" {
		header = lipgloss.JoinVertical(lipgloss.Left, header, errorStyle.Render(m.lastError))
	}

	var options []string
	for i, choice := range m.menuOptions {
		label := choice
		switch choice {
		case "Audio Source":
			label = fmt.Sprintf("Audio Source: %s", valueStyle.Render(m.selectedSource()))
		case "Chunk Duration":
			label = fmt.Sprintf("Chunk Duration: %s", valueStyle.Render(fmt.Sprintf("%ds", int(m.chunkDuration.Seconds()))))
		case "Start Session":
			if len(m.audioSources) == 0 {
				label = "Start Session (select audio source first)"
			}
		}

		if i == m.cursor {
			options = append(options, selectedMenuItemStyle.Render("▸ "+label))
			continue
		}
		options = append(options, menuItemStyle.Render("  "+label))
	}

	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		kbdStyle.Render("j/k or ↑/↓"), helpStyle.Render(" move   "),
		kbdStyle.Render("h/l or ←/→"), helpStyle.Render(" adjust   "),
		kbdStyle.Render("enter"), helpStyle.Render(" select   "),
		kbdStyle.Render("q"), helpStyle.Render(" quit"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		panelStyle.Render(strings.Join(options, "\n")),
		"",
		help,
	)
	return appFrameStyle.Render(content)
}

func (m *Model) renderRecording() string {
	status := statusRecordingStyle.Render(m.spinner.View() + " Recording")
	source := helpStyle.Render("Source: " + m.selectedSource())
	duration := helpStyle.Render(fmt.Sprintf("Chunk: %ds", int(m.chunkDuration.Seconds())))

	header := lipgloss.JoinHorizontal(lipgloss.Left, status, "   ", source, "   ", duration)
	transcriptPanel := transcriptPanelStyle.Render(transcriptStyle.Render(m.transcript.View()))
	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		kbdStyle.Render("↑/↓ or j/k"), helpStyle.Render(" scroll   "),
		kbdStyle.Render("s"), helpStyle.Render(" stop   "),
		kbdStyle.Render("q"), helpStyle.Render(" quit"),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		transcriptPanel,
		"",
		help,
	)
	return appFrameStyle.Render(content)
}

func (m *Model) selectedSource() string {
	if len(m.audioSources) > 0 && m.selectedSourceIndex >= 0 && m.selectedSourceIndex < len(m.audioSources) {
		return m.audioSources[m.selectedSourceIndex]
	}
	if !m.sourcesLoaded {
		return "Loading..."
	}
	return "No sources found"
}

func (m *Model) resizeTranscript() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	usableWidth := m.width - appFrameStyle.GetHorizontalFrameSize()
	if usableWidth < 42 {
		usableWidth = 42
	}

	usableHeight := m.height - appFrameStyle.GetVerticalFrameSize() - 8
	if usableHeight < 8 {
		usableHeight = 8
	}

	innerWidth := usableWidth - transcriptPanelStyle.GetHorizontalFrameSize()
	innerHeight := usableHeight - transcriptPanelStyle.GetVerticalFrameSize()
	if innerWidth < 20 {
		innerWidth = 20
	}
	if innerHeight < 5 {
		innerHeight = 5
	}

	m.transcript.Width = innerWidth
	m.transcript.Height = innerHeight
}
