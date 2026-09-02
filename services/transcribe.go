package services

import (
	"context"
	"sync"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
	"github.com/tuanta7/ekko/services/adapter/whisper"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type TranscribeService struct {
	mu  sync.Mutex
	ctx context.Context
	app *application.App

	scriber  *whisper.Scriber
	recorder *ffmpeg.Recorder

	sessions map[string]*Session
}

var (
	_ application.ServiceStartup  = (*TranscribeService)(nil)
	_ application.ServiceShutdown = (*TranscribeService)(nil)
)

func (t *TranscribeService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.app = application.Get()
	t.ctx = ctx
	t.recorder = ffmpeg.NewRecorder()
	t.sessions = make(map[string]*Session)
	return nil
}

func (t *TranscribeService) ServiceShutdown() error {
	t.mu.Lock()
	sessions := make([]*Session, 0, len(t.sessions))
	for _, sess := range t.sessions {
		sessions = append(sessions, sess)
	}
	t.mu.Unlock()

	for _, session := range sessions {
		session.Shutdown()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.scriber != nil {
		err := t.scriber.Close()
		t.scriber = nil
		return err
	}
	return nil
}
