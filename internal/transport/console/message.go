package console

import (
	tea "github.com/charmbracelet/bubbletea"
)

type sourcesLoadedMsg struct {
	sources []string
	err     error
}

type sessionEndMsg struct {
	err error
}

func (m *Model) sessionEnd(err error) tea.Cmd {
	return func() tea.Msg {
		return sessionEndMsg{
			err: err,
		}
	}
}
