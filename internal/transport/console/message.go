package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionEndMsg struct {
	Error error
}

func (m *Model) sessionEnd(err error) tea.Cmd {
	return func() tea.Msg {
		return sessionEndMsg{
			Error: err,
		}
	}
}

type transcriptMsg struct {
	chunk string
}

func (m *Model) waitForTranscript() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.ctx.Done():
			return sessionEndMsg{Error: nil}
		case chunk, ok := <-m.stream:
			if !ok {
				err := m.handler.Stop()
				if err == nil {
					err = fmt.Errorf("stream closed")
				}

				return m.sessionEnd(err)
			}
			return transcriptMsg{chunk: chunk}
		}
	}
}
