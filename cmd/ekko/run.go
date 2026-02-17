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
)

func RunCommand(zl *logger.Logger, recorder *ffmpeg.Recorder) *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run Ekko",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Value:   ":8080",
				Usage:   "Run Ekko in server mode and listen on the specified address",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    "model",
				Value:   "ggml-base",
				Usage:   "Specify the model to use",
				Aliases: []string{"m"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			modelName := cmd.String("model")
			scriber, err := whisper.NewClient(modelName)
			if err != nil {
				return fmt.Errorf("failed to create transcriber client: %v", err)
			}
			defer silent.Close(scriber)

			h := handler.NewHandler(recorder, scriber, zl)
			defer h.Close()

			if addr := cmd.String("server"); addr != "" {
				fmt.Println("Ekko is running in web mode")
				server := browser.NewServer(h)
				return server.Run(addr)
			}

			_, err = tea.NewProgram(console.NewModel(h)).Run()
			if err != nil {
				return fmt.Errorf("failed to run program %v", err)
			}

			return nil
		},
	}
}
