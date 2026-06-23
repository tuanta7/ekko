package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultSampleRate    = 16000
	DefaultFrameDuration = 100 * time.Millisecond
	DefaultBufferSeconds = 30
)

type Recorder struct{}

func NewRecorder() *Recorder {
	return &Recorder{}
}

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

	rows := strings.Split(output, "\n")
	sources := make([]string, len(rows))
	for i, row := range rows {
		parts := strings.Fields(row)
		sources[i] = strings.TrimSpace(parts[1])
	}

	return sources, nil
}

type RecordOptions struct {
	Source        string
	FrameDuration time.Duration
	BufferSeconds int
}

func (r *Recorder) Stream(ctx context.Context, options RecordOptions) (<-chan Frame, <-chan error, error) {
	source := strings.TrimSpace(options.Source)
	if source == "" {
		return nil, nil, errors.New("audio source is required")
	}

	frameDuration := options.FrameDuration
	if frameDuration <= 0 {
		frameDuration = DefaultFrameDuration
	}

	bufferSeconds := options.BufferSeconds
	if bufferSeconds <= 0 {
		bufferSeconds = DefaultBufferSeconds
	}

	frameSamples := int(frameDuration.Seconds() * float64(DefaultSampleRate))
	if frameSamples <= 0 {
		return nil, nil, fmt.Errorf("invalid frame duration %s", frameDuration)
	}

	bufferFrames := int(time.Duration(bufferSeconds) * time.Second / frameDuration)
	if bufferFrames < 1 {
		bufferFrames = 1
	}

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "pulse",
		"-i", source,
		"-ac", "1",
		"-ar", fmt.Sprintf("%d", DefaultSampleRate),
		"-f", "f32le",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	frames := make(chan Frame, bufferFrames)
	errs := make(chan error, 1)

	go func() {
		defer close(frames)
		defer close(errs)

		readErr := readFrames(ctx, stdout, frames, frameSamples)
		if readErr != nil && !errors.Is(readErr, context.Canceled) {
			_ = cmd.Process.Kill()
		}

		waitErr := cmd.Wait()
		if ctx.Err() != nil {
			return
		}

		if readErr != nil {
			errs <- readErr
			return
		}

		if waitErr != nil {
			message := strings.TrimSpace(stderr.String())
			if message != "" {
				errs <- fmt.Errorf("ffmpeg failed: %w: %s", waitErr, message)
				return
			}
			errs <- waitErr
		}
	}()

	return frames, errs, nil
}
