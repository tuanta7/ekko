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

func (ts *TranscribeService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.app = application.Get()
	ts.ctx = ctx
	ts.recorder = ffmpeg.NewRecorder()
	ts.sessions = make(map[string]*session.TranscribeSession)
}

func (ts *TranscribeService) ServiceShutdown() error {
	ts.mu.Lock()
	sessions := make([]*session.TranscribeSession, 0, len(ts.sessions))
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

	options.ModelName = normalizeModelName(options.ModelName)
	if options.Language == "" {
		options.Language = "auto"
	}

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

	if err := ts.ensureScriber(options.ModelName); err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(ts.context())
	frames, recorderErrs, err := ts.recorder.Stream(ctx, ffmpeg.RecordOptions{
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

	ts.mu.Lock()
	ts.sessions[id] = s
	ts.mu.Unlock()

	ts.emitState(id, "recording", "Recording started")
	go ts.runSession(ctx, s, frames, recorderErrs, options)

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

	select {
	case <-s.Done:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("timed out stopping transcription session")
	}
}

func (ts *TranscribeService) runSession(ctx context.Context, sess *session.TranscribeSession, frames <-chan ffmpeg.Frame, recorderErrs <-chan error, options StartOptions) {
	defer func() {
		ts.mu.Lock()
		delete(ts.sessions, sess.ID)
		ts.mu.Unlock()
		ts.emitState(sess.ID, "stopped", "Recording stopped")
		close(sess.Done)
	}()

	jobs := make(chan session.TranscribeJob, 4)
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		ts.runTranscriptionWorker(sess.ID, jobs, options)
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
				ts.enqueueChunks(ctx, jobs, chunker.Flush(), &chunkID)
				return
			}
			ts.enqueueChunks(ctx, jobs, chunker.AddFrame(frame), &chunkID)

		case err, ok := <-recorderErrs:
			if !ok {
				recorderErrs = nil
				continue
			}
			if ok && err != nil {
				ts.emitError(sess.ID, err)
				return
			}

		case <-ctx.Done():
			ts.enqueueChunks(context.Background(), jobs, chunker.Flush(), &chunkID)
			return
		}
	}
}

func (ts *TranscribeService) enqueueChunks(ctx context.Context, jobs chan<- session.TranscribeJob, chunks []session.AudioChunk, chunkID *int64) {
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

func (ts *TranscribeService) runTranscriptionWorker(sessionID string, jobs <-chan session.TranscribeJob, options StartOptions) {
	for job := range jobs {
		ts.transcribeJob(sessionID, job, options)
	}
}

func (ts *TranscribeService) transcribeJob(sessionID string, job session.TranscribeJob, options StartOptions) {
	ts.emitState(sessionID, "transcribing", "")

	segments, err := ts.scriber.Transcribe(job.Chunk.Samples, whisper.TranscribeOptions{
		Language:        options.Language,
		Threads:         options.Threads,
		Translate:       options.Translate,
		TokenTimestamps: job.Chunk.Final,
	})
	if err != nil {
		ts.emitError(sessionID, err)
		return
	}

	text := combineSegments(segments)
	if text == "" {
		return
	}

	ts.emitTranscript(TranscriptEvent{
		SessionID: sessionID,
		ChunkID:   job.ID,
		Text:      text,
		Final:     job.Chunk.Final,
		StartMs:   job.Chunk.Start.Milliseconds(),
		EndMs:     job.Chunk.End.Milliseconds(),
	})
	if job.Chunk.Final {
		ts.emitState(sessionID, "recording", "")
	}
}

func (ts *TranscribeService) ensureRecorder() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.recorder != nil {
		return
	}

	ts.recorder = ffmpeg.NewRecorder()
	if ts.sessions == nil {
		ts.sessions = make(map[string]*session.TranscribeSession)
	}
}

func (ts *TranscribeService) ensureScriber(modelName string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.scriber != nil && ts.model == modelName {
		return nil
	}
	if len(ts.sessions) > 0 {
		return errors.New("cannot change model while a transcription session is running")
	}
	if ts.scriber != nil {
		if err := ts.scriber.Close(); err != nil {
			return err
		}
		ts.scriber = nil
	}

	scriber, err := whisper.NewScriber(modelName)
	if err != nil {
		return err
	}

	ts.scriber = scriber
	ts.model = modelName
	return nil
}

func (ts *TranscribeService) context() context.Context {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.ctx != nil {
		return ts.ctx
	}
	return context.Background()
}

func (ts *TranscribeService) emitState(sessionID string, state string, message string) {
	ts.emit(event.StateEventName, event.StateEvent{
		SessionID: sessionID,
		State:     state,
		Message:   message,
	})
}

func (ts *TranscribeService) emitTranscript(event TranscriptEvent) {
	name := "transcribe:partial"
	if event.Final {
		name = "transcribe:final"
	}

	ts.emit(name, event)
}

func (ts *TranscribeService) emitError(sessionID string, err error) {
	if err == nil {
		return
	}

	ts.emit("transcribe:error", ErrorEvent{
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
