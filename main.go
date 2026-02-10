package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/ekko/internal/audio"
	"github.com/tuanta7/ekko/internal/manager"
	"github.com/tuanta7/ekko/internal/scribe"
	"github.com/tuanta7/ekko/internal/transport/console"
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

	recorder := audio.NewFFmpegRecorder()

	app := manager.NewHandler(recorder, s, fl)
	_, err = tea.NewProgram(console.NewModel(app, fl)).Run()
	if err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}
