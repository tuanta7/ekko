package whisper

import (
	"fmt"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type Scriber struct {
	model whisper.Model
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
