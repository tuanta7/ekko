package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

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

	sessions map[string]*TranscribeSession
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
	t.sessions = make(map[string]*TranscribeSession)
	return nil
}

func (t *TranscribeService) ServiceShutdown() error {
	t.mu.Lock()
	sessions := make([]*TranscribeSession, 0, len(t.sessions))
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

func (t *TranscribeService) ListSources() ([]string, error) {
	ctx, cancel := context.WithTimeout(t.ctx, 3*time.Second)
	defer cancel()

	return t.recorder.ListSources(ctx)
}

// initScriber initializes the Whisper scriber lazily.
func (t *TranscribeService) initScriber() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.scriber != nil {
		return nil
	}

	scriber, err := whisper.NewScriber()
	if err != nil {
		return err
	}

	t.scriber = scriber
	return nil
}

func (t *TranscribeService) Start(source string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		sources, err := t.ListSources()
		if err != nil {
			return "", err
		}
		if len(sources) == 0 {
			return "", errors.New("no audio sources available")
		}
		source = sources[0]
	}

	t.mu.Lock()
	if len(t.sessions) > 0 {
		t.mu.Unlock()
		return "", errors.New("a transcription session is already running")
	}
	t.mu.Unlock()

	if err := t.initScriber(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(t.ctx)
	frames, recorderErrs, err := t.recorder.Stream(ctx, source)
	if err != nil {
		cancel()
		return "", err
	}

	session := NewSession(cancel)

	t.mu.Lock()
	t.sessions[session.ID] = session
	t.mu.Unlock()

	// Notify listeners that audio recording has started for this session.
	t.emitState(session.ID, EventRecording, "Recording started")
	go t.runSession(ctx, session, frames, recorderErrs)

	return session.ID, nil
}

func (t *TranscribeService) Stop(sessionID string) error {
	t.mu.Lock()
	session, ok := t.sessions[sessionID]
	t.mu.Unlock()
	if !ok {
		return errors.New("transcription session not found")
	}

	session.Cancel()
	return nil
}
