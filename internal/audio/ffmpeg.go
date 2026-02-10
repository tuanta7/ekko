package audio

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type FFmpegRecorder struct {
	source string
}

func NewFFmpegRecorder() *FFmpegRecorder {
	return &FFmpegRecorder{}
}

func (r *FFmpegRecorder) ListSources(ctx context.Context) ([]string, error) {
	listCmd := exec.CommandContext(ctx, "pactl", "list", "sinks")

	var buf bytes.Buffer
	listCmd.Stdout = &buf

	err := listCmd.Run()
	if err != nil {
		return nil, err
	}

	return strings.Split(buf.String(), "\n"), nil
}

func (r *FFmpegRecorder) SetSource(source string) {
	r.source = source
}

func (r *FFmpegRecorder) Record(ctx context.Context, duration time.Duration, outputFile string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "pulse",
		"-i", r.source,
		"-t", fmt.Sprintf("%0.2f", duration.Seconds()),
		"-ar", "16000", // 16kHz sample rate
		"-ac", "1", // mono audio
		"-y", // overwrite output file
		outputFile,
	)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
