package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()

	gc, err := NewGeminiClient(ctx, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("failed to create Gemini client: %v\n", err)
	}

	r := NewRecorder()
	err = r.Record(10*time.Second, "output.wav")
	if err != nil {
		log.Fatalf("failed to record audio: %v\n", err)
	}

	contents, err := gc.NewContentsFromAudio(ctx, "output.wav")
	if err != nil {
		log.Fatalf("failed to create contents from audio: %v\n", err)
	}

	text, err := gc.Transcribe(ctx, contents)
	if err != nil {
		log.Fatalf("failed to transcribe audio: %v\n", err)
	}

	if text != "" {
		log.Println(text)
	}
}
