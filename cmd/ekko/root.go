package main

import (
	"context"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/internal/handler"
	"github.com/tuanta7/ekko/internal/whisper"
	"github.com/tuanta7/ekko/pkg/logger"
	"github.com/tuanta7/ekko/pkg/silent"
	"github.com/urfave/cli/v3"
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

	h := handler.NewHandler(recorder, scriber, zl)
	defer h.Close()

	cmd := &cli.Command{
		Name:  "ekko",
		Usage: "Ekko is a CLI tool for transcribing audio files using whisper-cpp.",
		Commands: []*cli.Command{
			RunCommand(h),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
