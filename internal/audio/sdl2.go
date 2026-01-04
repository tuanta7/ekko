package audio

import (
	"context"
	"time"
)

type SDL2Recorder struct{}

func NewSDL2Recorder() *SDL2Recorder {
	return &SDL2Recorder{}
}

func (r *SDL2Recorder) Record(ctx context.Context, duration time.Duration, source, outputFile string) error {
	return nil
}
