package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/internal/handler"
	"github.com/tuanta7/ekko/internal/transport/browser"
	"github.com/tuanta7/ekko/internal/transport/console"
	"github.com/tuanta7/ekko/internal/whisper"
	"github.com/tuanta7/ekko/pkg/logger"
	"github.com/tuanta7/ekko/pkg/silent"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
)

func RunCommand(zl *logger.Logger, recorder *ffmpeg.Recorder) *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run Ekko (default: TUI mode)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "web",
				Usage:   "Enable web/server mode",
				Aliases: []string{"w"},
			},
			&cli.StringFlag{
				Name:    "addr",
				Value:   ":8080",
				Usage:   "Address to listen on when running in web mode (only used with --web)",
				Aliases: []string{"a"},
			},
			&cli.StringFlag{
				Name:    "model",
				Value:   "ggml-base",
				Usage:   "Specify the Whisper GGML model name to use",
				Aliases: []string{"m"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			modelName := cmd.String("model")
			scriber, err := whisper.NewClient(modelName)
			if err != nil {
				return fmt.Errorf("failed to create transcriber client: %w", err)
			}
			defer silent.Close(scriber)

			h := handler.NewHandler(recorder, scriber, zl)
			defer h.Close()

			if cmd.Bool("web") {
				addr := cmd.String("addr")
				zl.Info("starting web mode", zap.String("addr", addr))
				server := browser.NewServer(h)
				return server.Run(addr)
			}

			_, err = tea.NewProgram(console.NewModel(h)).Run()
			if err != nil {
				return fmt.Errorf("failed to run TUI: %w", err)
			}

			return nil
		},
	}
}
