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

// ListSources and inputArgs are per-OS: PulseAudio on Linux, AVFoundation on
// macOS. See recorder_linux.go / recorder_darwin.go.

type RecordOptions struct {
	Source        string
	FrameDuration time.Duration
	BufferSeconds int
}

func (r *Recorder) Stream(ctx context.Context, source string) (<-chan Frame, <-chan error, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, nil, errors.New("audio source is required")
	}

	frameDuration := DefaultFrameDuration
	bufferSeconds := DefaultBufferSeconds

	//  number of audio samples in one frame
	//  0.1 seconds × 16,000 samples/second = 1,600 samples
	frameSamples := int(frameDuration.Seconds() * float64(DefaultSampleRate))
	if frameSamples <= 0 {
		return nil, nil, fmt.Errorf("invalid frame duration %s", frameDuration)
	}

	bufferFrames := int(time.Duration(bufferSeconds) * time.Second / frameDuration)
	if bufferFrames < 1 {
		bufferFrames = 1
	}

	// FFmpeg emits a continuous PCM stream; readFrames applies the frame
	// boundaries using frameSamples rather than relying on FFmpeg packets.
	args := append([]string{"-hide_banner", "-loglevel", "error"}, inputArgs(source)...)
	args = append(args,
		"-ac", "1", // Downmix the input to mono.
		"-ar", fmt.Sprintf("%d", DefaultSampleRate),
		"-f", "f32le", // Emit headerless, little-endian float32 PCM.
		"pipe:1", // Stream PCM through stdout.
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

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
			if message := strings.TrimSpace(stderr.String()); message != "" {
				errs <- fmt.Errorf("ffmpeg failed: %w: %s", waitErr, message)
				return
			}
			errs <- waitErr
		}
	}()

	return frames, errs, nil
}
