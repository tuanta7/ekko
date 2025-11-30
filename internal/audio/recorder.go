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

type Recorder struct {
	storage []any
}

func NewRecorder() *Recorder {
	return &Recorder{}
}

func (r *Recorder) Record(ctx context.Context, duration time.Duration, source, outputFile string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "pulse",
		"-i", source,
		"-t", fmt.Sprintf("%0.2f", duration.Seconds()),
		"-y", // overwrite output file
		outputFile,
	)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (r *Recorder) GetSource(ctx context.Context) (string, error) {
	pr, pw := io.Pipe()
	defer pr.Close()

	var buf bytes.Buffer

	grepCmd := exec.CommandContext(ctx, "grep", ".monitor") // system audio
	grepCmd.Stdout = &buf
	grepCmd.Stdin = pr // receives data from the pipe's read-end

	listCmd := exec.CommandContext(ctx, "pactl", "list", "sinks")
	listCmd.Stdout = pw // writes into the pipe's write-end

	if err := grepCmd.Start(); err != nil {
		_ = pw.Close()
		return "", err
	}

	if err := listCmd.Run(); err != nil {
		_ = pw.Close()
		_ = grepCmd.Wait()
		return "", err
	}
	// close writer to signal EOF to grep
	_ = pw.Close()

	if err := grepCmd.Wait(); err != nil {
		return "", err
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		return "", errors.New("no monitor sink found")
	}

	line := output
	if i := strings.IndexByte(output, '\n'); i >= 0 {
		// use the first matching line
		line = output[:i]
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to parse pactl output: %q", line)
	}

	if src := strings.TrimSpace(parts[1]); src != "" {
		return src, nil
	}

	return "", errors.New("no monitor sink found")
}
