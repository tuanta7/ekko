package services

import (
	"context"
	"time"
)

func (t *TranscribeService) ListSources() ([]string, error) {
	t.ensureRecorder()

	ctx, cancel := context.WithTimeout(t.context(), 3*time.Second)
	defer cancel()

	return t.recorder.ListSources(ctx)
}
