package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
	"github.com/tuanta7/ekko/services/adapter/whisper"
	"github.com/tuanta7/ekko/services/event"
	"github.com/tuanta7/ekko/services/session"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	defaultModelName = "ggml-base"
)

type TranscribeService struct {
	mu sync.Mutex

	app *application.App
	ctx context.Context

	model    string
	scriber  *whisper.Scriber
	recorder *ffmpeg.Recorder

	sessions map[string]*session.TranscribeSession
}

func (t *TranscribeService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.app = application.Get()
	t.ctx = ctx
	t.recorder = ffmpeg.NewRecorder()
	t.sessions = make(map[string]*session.TranscribeSession)
}

func (t *TranscribeService) ServiceShutdown() error {
	t.mu.Lock()
	sessions := make([]*session.TranscribeSession, 0, len(t.sessions))
	for _, sess := range t.sessions {
		sessions = append(sessions, sess)
	}
	t.mu.Unlock()

	for _, sess := range sessions {
		sess.Shutdown()
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

func (t *TranscribeService) Start(options StartOptions) (string, error) {
	t.ensureRecorder()

	options.ModelName = normalizeModelName(options.ModelName)
	if options.Language == "" {
		options.Language = "auto"
	}

	source := strings.TrimSpace(options.Source)
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

	if err := t.ensureScriber(options.ModelName); err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(t.context())
	frames, recorderErrs, err := t.recorder.Stream(ctx, ffmpeg.RecordOptions{
		Source:        source,
		FrameDuration: ffmpeg.DefaultFrameDuration,
		BufferSeconds: ffmpeg.DefaultBufferSeconds,
	})
	if err != nil {
		cancel()
		return "", err
	}

	s := session.NewSession(cancel)
	id := s.ID

	t.mu.Lock()
	t.sessions[id] = s
	t.mu.Unlock()

	t.emitState(id, "recording", "Recording started")
	go t.runSession(ctx, s, frames, recorderErrs, options)

	return id, nil
}

func (t *TranscribeService) Stop(sessionID string) error {
	t.mu.Lock()
	s, ok := t.sessions[sessionID]
	t.mu.Unlock()
	if !ok {
		return errors.New("transcription session not found")
	}

	s.Cancel()

	select {
	case <-s.Done:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("timed out stopping transcription session")
	}
}

func (t *TranscribeService) runSession(ctx context.Context, sess *session.TranscribeSession, frames <-chan ffmpeg.Frame, recorderErrs <-chan error, options StartOptions) {
	defer func() {
		t.mu.Lock()
		delete(t.sessions, sess.ID)
		t.mu.Unlock()
		t.emitState(sess.ID, "stopped", "Recording stopped")
		close(sess.Done)
	}()

	jobs := make(chan session.TranscribeJob, 4)
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		t.runTranscriptionWorker(sess.ID, jobs, options)
	}()
	defer func() {
		close(jobs)
		workers.Wait()
	}()

	chunker := session.NewAudioChunker()

	var chunkID int64
	for {
		select {
		case frame, ok := <-frames:
			if !ok {
				t.enqueueChunks(ctx, jobs, chunker.Flush(), &chunkID)
				return
			}
			t.enqueueChunks(ctx, jobs, chunker.AddFrame(frame), &chunkID)

		case err, ok := <-recorderErrs:
			if !ok {
				recorderErrs = nil
				continue
			}
			if ok && err != nil {
				t.emitError(sess.ID, err)
				return
			}

		case <-ctx.Done():
			t.enqueueChunks(context.Background(), jobs, chunker.Flush(), &chunkID)
			return
		}
	}
}

func (s *TranscribeService) enqueueChunks(ctx context.Context, jobs chan<- session.TranscribeJob, chunks []session.AudioChunk, chunkID *int64) {
	for _, chunk := range chunks {
		*chunkID = *chunkID + 1
		job := session.TranscribeJob{ID: *chunkID, Chunk: chunk}
		if chunk.Final {
			select {
			case jobs <- job:
			case <-ctx.Done():
				return
			}
			continue
		}

		select {
		case jobs <- job:
		default:
		}
	}
}

func (s *TranscribeService) runTranscriptionWorker(sessionID string, jobs <-chan session.TranscribeJob, options StartOptions) {
	for job := range jobs {
		s.transcribeJob(sessionID, job, options)
	}
}

func (s *TranscribeService) transcribeJob(sessionID string, job session.TranscribeJob, options StartOptions) {
	s.emitState(sessionID, "transcribing", "")

	segments, err := s.scriber.Transcribe(job.Chunk.Samples, whisper.TranscribeOptions{
		Language:        options.Language,
		Threads:         options.Threads,
		Translate:       options.Translate,
		TokenTimestamps: job.Chunk.Final,
	})
	if err != nil {
		s.emitError(sessionID, err)
		return
	}

	text := combineSegments(segments)
	if text == "" {
		return
	}

	s.emitTranscript(TranscriptEvent{
		SessionID: sessionID,
		ChunkID:   job.ID,
		Text:      text,
		Final:     job.Chunk.Final,
		StartMs:   job.Chunk.Start.Milliseconds(),
		EndMs:     job.Chunk.End.Milliseconds(),
	})
	if job.Chunk.Final {
		s.emitState(sessionID, "recording", "")
	}
}

func (s *TranscribeService) ensureRecorder() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.recorder != nil {
		return
	}

	s.recorder = ffmpeg.NewRecorder()
	if s.sessions == nil {
		s.sessions = make(map[string]*session.TranscribeSession)
	}
}

func (s *TranscribeService) ensureScriber(modelName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.scriber != nil && s.model == modelName {
		return nil
	}
	if len(s.sessions) > 0 {
		return errors.New("cannot change model while a transcription session is running")
	}
	if s.scriber != nil {
		if err := s.scriber.Close(); err != nil {
			return err
		}
		s.scriber = nil
	}

	scriber, err := whisper.NewScriber(modelName)
	if err != nil {
		return err
	}

	s.scriber = scriber
	s.model = modelName
	return nil
}

func (s *TranscribeService) context() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx != nil {
		return s.ctx
	}
	return context.Background()
}

func (s *TranscribeService) emitState(sessionID string, state string, message string) {
	s.emit(event.StateEventName, event.StateEvent{
		SessionID: sessionID,
		State:     state,
		Message:   message,
	})
}

func (s *TranscribeService) emitTranscript(event TranscriptEvent) {
	name := "transcribe:partial"
	if event.Final {
		name = "transcribe:final"
	}

	s.emit(name, event)
}

func (s *TranscribeService) emitError(sessionID string, err error) {
	if err == nil {
		return
	}

	s.emit("transcribe:error", ErrorEvent{
		SessionID: sessionID,
		Message:   err.Error(),
	})
}

func normalizeModelName(modelName string) string {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return defaultModelName
	}

	if strings.HasPrefix(modelName, "ggml-") {
		return modelName
	}

	return "ggml-" + modelName
}

func combineSegments(segments []whisper.Segment) string {
	var parts []string
	for _, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}
