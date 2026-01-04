package audio

import (
	"context"
	"time"
)

type Recorder interface {
	Record(ctx context.Context, duration time.Duration, outputFile string) error
}
