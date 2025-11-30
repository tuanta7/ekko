package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/transcript/internal/audio"
	"github.com/tuanta7/transcript/internal/core"
	"github.com/tuanta7/transcript/internal/transcriber"
	"github.com/tuanta7/transcript/internal/ui"
)

func main() {
	ctx := context.Background()

	mode := transcriber.Mode(os.Getenv("TRANSCRIBER_MODE"))
	gc, err := transcriber.NewClient(ctx, mode, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		fmt.Printf("Failed to create transcriber client: %v", err)
		os.Exit(1)
	}
	defer gc.Close()

	recorder := audio.NewRecorder()
	app := core.NewApplication(recorder, gc)

	model := ui.NewModel(app)
	_, err = tea.NewProgram(model).Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
