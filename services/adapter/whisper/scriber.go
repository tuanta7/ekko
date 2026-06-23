package whisper

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type Scriber struct {
	model whisper.Model
	mu    sync.Mutex
}

type TranscribeOptions struct {
	Language        string
	Threads         uint
	Translate       bool
	TokenTimestamps bool
	InitialPrompt   string
}

type Segment struct {
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
	Text  string        `json:"text"`
}

func NewScriber(modelName string) (*Scriber, error) {
	model, err := whisper.New(fmt.Sprintf("ggml/%s.bin", modelName))
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

func (s *Scriber) DownloadModel(modelName string) error {
	return nil
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

	if options.Language != "" && options.Language != "auto" {
		if err := ctx.SetLanguage(options.Language); err != nil {
			return nil, err
		}
	}

	if options.Threads > 0 {
		ctx.SetThreads(options.Threads)
	}

	ctx.SetTranslate(options.Translate)
	ctx.SetTokenTimestamps(options.TokenTimestamps)
	if options.InitialPrompt != "" {
		ctx.SetInitialPrompt(options.InitialPrompt)
	}

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
