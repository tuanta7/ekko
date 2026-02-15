package whisper

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type Client struct {
	model whisper.Model
	ctx   whisper.Context
}

func NewClient() (*Client, error) {
	model, err := whisper.New(fmt.Sprintf("ggml/%s.bin", os.Getenv("MODEL_NAME")))
	if err != nil {
		return nil, err
	}

	modelContext, err := model.NewContext()
	if err != nil {
		return nil, err
	}

	fmt.Println("Whisper model loaded successfully!")
	return &Client{
		model: model,
		ctx:   modelContext,
	}, nil
}

func (l *Client) Close() error {
	return l.model.Close()
}

func (l *Client) ResetContext() error {
	modelContext, err := l.model.NewContext()
	if err != nil {
		return err
	}

	modelContext.SetTemperature(0.5)
	l.ctx = modelContext
	return nil
}

var whisperTagRE = regexp.MustCompile(`\[_BEG_]\s*|\[_EOT_]\s*|\[_TT_\d+]\s*`)

func (l *Client) Transcribe(ctx context.Context, data []float32, out chan string) error {
	return l.ctx.Process(data, nil, func(segment whisper.Segment) {
		text := whisperTagRE.ReplaceAllString(segment.Text, "")
		if text != "" {
			select {
			case <-ctx.Done():
			case out <- text:
			}
		}
	}, nil)
}
