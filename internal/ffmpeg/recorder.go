package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
)

type Recorder struct {
	source string
}

func NewRecorder() (*Recorder, error) {
	if err := os.RemoveAll("tmp"); err != nil {
		return nil, err
	}

	if err := os.Mkdir("tmp", 0755); err != nil {
		return nil, err
	}

	return &Recorder{}, nil
}

func (r *Recorder) SetSource(source string) {
	r.source = source
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

func (r *Recorder) Record(ctx context.Context, duration time.Duration) ([]float32, error) {
	filePath := fmt.Sprintf("tmp/whisper-%d.wav", time.Now().UnixNano())

	cmd := exec.CommandContext(ctx, "ffmpeg", "-f", "pulse",
		"-i", r.source,
		"-t", fmt.Sprintf("%0.2f", duration.Seconds()),
		"-ar", "16000", // 16kHz sample rate
		"-ac", "1", // mono audio
		"-y", // overwrite output file
		filePath,
	)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, err
	} else if dec.SampleRate != whisper.SampleRate {
		return nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	} else if dec.NumChans != 1 {
		return nil, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	}

	return buf.AsFloat32Buffer().Data, nil
}
