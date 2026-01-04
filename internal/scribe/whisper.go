package scribe

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
)

type WhisperClient struct {
	model whisper.Model
	ctx   whisper.Context
}

func NewLocalClient() (*WhisperClient, error) {
	model, err := whisper.New(fmt.Sprintf("models/%s.bin", os.Getenv("MODEL_NAME")))
	if err != nil {
		return nil, err
	}

	modelContext, err := model.NewContext()
	if err != nil {
		return nil, err
	}

	fmt.Println("Whisper model loaded successfully!")
	return &WhisperClient{
		model: model,
		ctx:   modelContext,
	}, nil
}

func (l *WhisperClient) Close() error {
	return l.model.Close()
}

func (l *WhisperClient) ResetContext() error {
	modelContext, err := l.model.NewContext()
	if err != nil {
		return err
	}

	modelContext.SetTemperature(0.5)
	modelContext.SetInitialPrompt(InitialPrompts)

	l.ctx = modelContext
	return nil
}

var whisperTagRE = regexp.MustCompile(`\[_BEG_]\s*|\[_EOT_]\s*|\[_TT_\d+]\s*`)

func (l *WhisperClient) Transcribe(ctx context.Context, audioPath string, out chan string) error {
	f, err := os.Open(audioPath)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return err
	} else if dec.SampleRate != whisper.SampleRate {
		return fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	} else if dec.NumChans != 1 {
		return fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	}

	cb := func(segment whisper.Segment) {
		text := whisperTagRE.ReplaceAllString(segment.Text, "")
		if text != "" {
			select {
			case <-ctx.Done():
			case out <- text:
			}
		}
	}

	data := buf.AsFloat32Buffer().Data
	return l.ctx.Process(data, nil, cb, nil)
}
