package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
)

// ListSources returns the PulseAudio/PipeWire source names, including the
// `.monitor` sinks that carry what the machine is playing.
func (r *Recorder) ListSources(ctx context.Context) ([]string, error) {
	listCmd := exec.CommandContext(ctx, "pactl", "list", "sources", "short")

	var buf bytes.Buffer
	listCmd.Stdout = &buf

	if err := listCmd.Run(); err != nil {
		return nil, err
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		return nil, errors.New("no monitor sink found")
	}

	var sources []string
	for _, row := range strings.Split(output, "\n") {
		parts := strings.Fields(row)
		if len(parts) < 2 {
			continue
		}
		sources = append(sources, parts[1])
	}

	return sources, nil
}

func inputArgs(source string) []string {
	return []string{"-f", "pulse", "-i", source}
}
