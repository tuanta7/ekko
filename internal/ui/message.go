package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionEndMsg struct {
	Timestamp time.Time
	Filename  string
	Error     error
}

func (m *Model) sessionEnd() tea.Cmd {
	filename, err := m.app.Stop()
	return func() tea.Msg {
		return sessionEndMsg{
			Timestamp: time.Now(),
			Filename:  filename,
			Error:     err,
		}
	}
}

type transcriptChunkMsg struct {
	Text string
}

func (m *Model) waitForTranscript() tea.Cmd {
	return func() tea.Msg {
		text, ok := <-m.stream
		if !ok {
			_, _ = m.app.Stop()
			return sessionEndMsg{
				Timestamp: time.Now(),
				Error:     fmt.Errorf("stream closed"),
			}
		}
		return transcriptChunkMsg{Text: text}
	}
}
