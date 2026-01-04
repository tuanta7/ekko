package logger

import (
	tea "github.com/charmbracelet/bubbletea"
)

type FileLogger struct {
	filePath string
}

func NewLogger(filePath string) *FileLogger {
	return &FileLogger{
		filePath: filePath,
	}
}

func (l *FileLogger) Debug(msg string) error {
	return l.logToFile(msg, "debug")
}

func (l *FileLogger) Info(msg string) error {
	return l.logToFile(msg, "info")
}

func (l *FileLogger) Warn(msg string) error {
	return l.logToFile(msg, "warn")
}

func (l *FileLogger) Error(msg string) error {
	return l.logToFile(msg, "error")
}

func (l *FileLogger) logToFile(msg string, prefix string) error {
	file, err := tea.LogToFile(l.filePath, prefix)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(msg))
	return err
}
