package logger

import "github.com/wailsapp/wails/v3/pkg/application"

type FileLogger struct {
	app *application.App
}

func (fl *FileLogger) Info(msg string) {
	fl.app.Logger.Info(msg)
}
