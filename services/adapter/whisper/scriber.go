package whisper

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

const modelPath = "ggml/ggml-tiny.en-q5_1.bin"

type Scriber struct {
	model whisper.Model
	mu    sync.Mutex
}

type TranscribeOptions struct {
	TokenTimestamps bool
}

type Segment struct {
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
	Text  string        `json:"text"`
}

func NewScriber() (*Scriber, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, err
	}

	return &Scriber{
		model: model,
	}, nil
}

func (s *Scriber) Close() error {
	return s.model.Close()
}

func (s *Scriber) Transcribe(samples []float32, options TranscribeOptions) ([]Segment, error) {
	if len(samples) == 0 {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, err := s.model.NewContext()
	if err != nil {
		return nil, err
	}

	ctx.SetTokenTimestamps(options.TokenTimestamps)

	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return nil, err
	}

	var segments []Segment
	for {
		segment, err := ctx.NextSegment()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}

		segments = append(segments, Segment{
			Start: segment.Start,
			End:   segment.End,
			Text:  text,
		})
	}

	return segments, nil
}

func CombineSegments(segments []Segment) string {
	var parts []string
	for _, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}
