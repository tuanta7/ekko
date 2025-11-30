package transcriber

import (
	"context"
	"io"
	"log"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type WhisperClient struct {
	model whisper.Model
	ctx   whisper.Context
}

func NewLocalClient() *WhisperClient {
	model, err := whisper.New("models/ggml-base.bin")
	if err != nil {
		log.Fatal(err)
	}

	return &WhisperClient{
		model: model,
	}
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
	err := l.ctx.Process(nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
