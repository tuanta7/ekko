package audio

import (
	"context"
	"time"
)

type Recorder interface {
	ListSources(ctx context.Context) ([]string, error)
	SetSource(source string)
	Record(ctx context.Context, duration time.Duration, outputFile string) error
}
