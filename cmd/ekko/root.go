package main

import (
	"context"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/pkg/logger"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
)

func main() {
	zl, err := logger.New(zap.DebugLevel)
	if err != nil {
		// Can't use project logger if it failed to initialise; fall back to stderr.
		_, _ = os.Stderr.WriteString("fatal: failed to create logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	recorder, err := ffmpeg.NewRecorder()
	if err != nil {
		zl.Fatal("failed to create recorder", zap.Error(err))
	}

	cmd := &cli.Command{
		Name:  "ekko",
		Usage: "Ekko is a near real-time desktop audio transcription tool.",
		Commands: []*cli.Command{
			RunCommand(zl, recorder),
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			return nil
		},
	}

	if err = cmd.Run(context.Background(), os.Args); err != nil {
		zl.Fatal("command failed", zap.Error(err))
	}
}
