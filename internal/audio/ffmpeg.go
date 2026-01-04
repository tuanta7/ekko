package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

type FFmpegRecorder struct {
	source string
}

func NewFFmpegRecorder(ctx context.Context) (*FFmpegRecorder, error) {
	listCmd := exec.CommandContext(ctx, "pactl", "list", "sinks")
	grepCmd := exec.CommandContext(ctx, "grep", ".monitor") // system audio

	var buf bytes.Buffer
	pr, pw := io.Pipe()
	defer pr.Close()

	grepCmd.Stdout = &buf
	grepCmd.Stdin = pr
	listCmd.Stdout = pw

	if err := grepCmd.Start(); err != nil {
		_ = pw.Close()
		return nil, err
	}

	if err := listCmd.Run(); err != nil {
		_ = pw.Close()
		_ = grepCmd.Wait()
		return nil, err
	}

	_ = pw.Close() // close writer to signal EOF to grep
	if err := grepCmd.Wait(); err != nil {
		return nil, err
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		return nil, errors.New("no monitor sink found")
	}

	line := output
	if i := strings.IndexByte(output, '\n'); i >= 0 {
		// use the first matching line
		line = output[:i]
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("failed to parse pactl output: %q", line)
	}

	src := strings.TrimSpace(parts[1])
	if src == "" {
		return nil, errors.New("no monitor sink found")
	}

	return &FFmpegRecorder{src}, nil
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
