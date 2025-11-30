package transcriber

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
)

type WhisperClient struct {
	model whisper.Model
	ctx   whisper.Context
}

func NewLocalClient() (*WhisperClient, error) {
	model, err := whisper.New("models/ggml-base.bin")
	if err != nil {
		return nil, err
	}

	modelContext, err := model.NewContext()
	if err != nil {
		return nil, err
	}

	fmt.Println("Whisper model loaded successfully!")
	fmt.Print("\n\n")

	return &WhisperClient{
		model: model,
		ctx:   modelContext,
	}, nil
}

func (l *WhisperClient) Close() error {
	return l.model.Close()
}

func (l *WhisperClient) ResetContext(_ context.Context) error {
	modelContext, err := l.model.NewContext()
	if err != nil {
		return err
	}

	l.ctx = modelContext
	return nil
}

func (l *WhisperClient) Transcribe(ctx context.Context, audioPath string) (io.ReadCloser, error) {
	f, err := os.Open(audioPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data []float32
	dec := wav.NewDecoder(f)
	if buf, err := dec.FullPCMBuffer(); err != nil {
		return nil, err
	} else if dec.SampleRate != whisper.SampleRate {
		return nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	} else if dec.NumChans != 1 {
		return nil, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	} else {
		data = buf.AsFloat32Buffer().Data
	}

	r, w := io.Pipe()
	cb := func(segment whisper.Segment) {
		for _, token := range segment.Tokens {
			select {
			case <-ctx.Done():
				return
			default:
				if _, err = w.Write([]byte(token.Text)); err != nil {
					_ = w.CloseWithError(err)
					return
				}
			}
		}
	}

	go func() {
		defer w.Close()
		if err = l.ctx.Process(data, nil, cb, nil); err != nil {
			_ = w.CloseWithError(err)
		}
	}()

	return r, nil
}
