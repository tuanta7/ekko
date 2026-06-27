package services

import (
	"context"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
)

func (ts *TranscribeService) ensureRecorder() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.recorder != nil {
		return
	}

	ts.recorder = ffmpeg.NewRecorder()
	if ts.sessions == nil {
		ts.sessions = make(map[string]*TranscribeSession)
	}
}

func (ts *TranscribeService) ListSources() ([]string, error) {
	ts.ensureRecorder()

	ctx, cancel := context.WithTimeout(ts.context(), 3*time.Second)
	defer cancel()

	return ts.recorder.ListSources(ctx)
}
