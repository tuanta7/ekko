package services

import (
	"context"
	"time"
)

func (ts *TranscribeService) ListSources() ([]string, error) {
	ts.ensureRecorder()

	ctx, cancel := context.WithTimeout(ts.context(), 3*time.Second)
	defer cancel()

	return ts.recorder.ListSources(ctx)
}
