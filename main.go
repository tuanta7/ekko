package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/internal/handler"
	"github.com/tuanta7/ekko/internal/transport/console"
	"github.com/tuanta7/ekko/internal/whisper"
	"github.com/tuanta7/ekko/pkg/logger"
	"github.com/tuanta7/ekko/pkg/silent"
	"go.uber.org/zap"
)

func main() {
	scriber, err := whisper.NewClient()
	if err != nil {
		log.Fatalf("Failed to create transcriber client: %v", err)
	}
	defer silent.Close(scriber)

	zl, err := logger.New(zap.DebugLevel)
	silent.PanicOnErr(err, "Failed to create logger")

	recorder, err := ffmpeg.NewRecorder()
	silent.PanicOnErr(err, "Failed to create recorder")

	app := handler.NewHandler(recorder, scriber, zl)
	_, err = tea.NewProgram(console.NewModel(app)).Run()
	silent.PanicOnErr(err, "Failed to run program")
}
