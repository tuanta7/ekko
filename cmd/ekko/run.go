package main

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tuanta7/ekko/internal/handler"
	"github.com/tuanta7/ekko/internal/transport/console"
	"github.com/urfave/cli/v3"
)

func RunCommand(h *handler.Handler) *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run Ekko",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "web",
				Value:   false,
				Usage:   "Run Ekko in web mode",
				Aliases: []string{"w"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("web") {
				fmt.Println("Ekko is running in web mode")
				return nil
			}

			_, err := tea.NewProgram(console.NewModel(h)).Run()
			if err != nil {
				return fmt.Errorf("failed to run program %v", err)
			}

			return nil
		},
	}
}
