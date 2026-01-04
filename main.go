package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/ekko/internal/audio"
	"github.com/tuanta7/ekko/internal/handler"
	"github.com/tuanta7/ekko/internal/scribe"
	"github.com/tuanta7/ekko/internal/ui"
	"github.com/tuanta7/ekko/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	mode := scribe.Mode(os.Getenv("TRANSCRIBER_MODE"))
	s, err := scribe.NewScribe(ctx, mode, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to create transcriber client: %v", err)
	}
	defer func(scriber scribe.Scriber) {
		if err = scriber.Close(); err != nil {
			fmt.Printf("Failed to close transcriber client: %v", err)
		}
	}(s)

	recorder, err := audio.NewFFmpegRecorder(ctx)
	fl := logger.NewLogger(fmt.Sprintf("./logs/%s.log", time.Now().Format("2006-01-02_15-04-05")))
	app := handler.NewHandler(recorder, s, fl)
	_, err = tea.NewProgram(ui.NewModel(app, fl)).Run()
	if err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
