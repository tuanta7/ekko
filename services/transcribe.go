package services

import (
	"context"
	"errors"
	"strings"
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

	sessions map[string]*TranscribeSession
}

var (
	_ application.ServiceStartup  = (*TranscribeService)(nil)
	_ application.ServiceShutdown = (*TranscribeService)(nil)
)

func (ts *TranscribeService) context() context.Context {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.ctx != nil {
		return ts.ctx
	}
	return context.Background()
}

func (ts *TranscribeService) ensureScriber() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.scriber != nil {
		return nil
	}

	scriber, err := whisper.NewScriber()
	if err != nil {
		return err
	}

	ts.scriber = scriber
	return nil
}

func (ts *TranscribeService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.app = application.Get()
	ts.ctx = ctx
	ts.recorder = ffmpeg.NewRecorder()
	ts.sessions = make(map[string]*TranscribeSession)

	return nil
}

func (ts *TranscribeService) ServiceShutdown() error {
	ts.mu.Lock()
	sessions := make([]*TranscribeSession, 0, len(ts.sessions))
	for _, sess := range ts.sessions {
		sessions = append(sessions, sess)
	}
	ts.mu.Unlock()

	for _, sess := range sessions {
		sess.Shutdown()
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.scriber != nil {
		err := ts.scriber.Close()
		ts.scriber = nil
		return err
	}
	return nil
}

func (ts *TranscribeService) Start(options StartOptions) (string, error) {
	ts.ensureRecorder()

	source := strings.TrimSpace(options.Source)
	if source == "" {
		sources, err := ts.ListSources()
		if err != nil {
			return "", err
		}
		if len(sources) == 0 {
			return "", errors.New("no audio sources available")
		}
		source = sources[0]
	}

	ts.mu.Lock()
	if len(ts.sessions) > 0 {
		ts.mu.Unlock()
		return "", errors.New("a transcription session is already running")
	}
	ts.mu.Unlock()

	if err := ts.ensureScriber(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(ts.context())
	frames, recorderErrs, err := ts.recorder.Stream(ctx, source)
	if err != nil {
		cancel()
		return "", err
	}

	s := NewSession(cancel)
	id := s.ID

	ts.mu.Lock()
	ts.sessions[id] = s
	ts.mu.Unlock()

	ts.emitState(id, "recording", "Recording started")
	go ts.runSession(ctx, s, frames, recorderErrs)

	return id, nil
}

func (ts *TranscribeService) Stop(sessionID string) error {
	ts.mu.Lock()
	s, ok := ts.sessions[sessionID]
	ts.mu.Unlock()
	if !ok {
		return errors.New("transcription session not found")
	}

	s.Cancel()
	return nil
}
