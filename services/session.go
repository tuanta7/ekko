package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
	"github.com/tuanta7/ekko/services/chunker"
)

type Job struct {
	ID    int64
	Chunk chunker.AudioChunk
}

type TranscribeSession struct {
	ID     string
	Cancel context.CancelFunc
	Done   chan struct{}
}

func NewSession(cancel context.CancelFunc) *TranscribeSession {
	return &TranscribeSession{
		ID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		Cancel: cancel,
		Done:   make(chan struct{}),
	}
}

func (t *TranscribeSession) Shutdown() {
	t.Cancel()
	<-t.Done
}

type StartOptions struct {
	Source string `json:"source"`
}

func (ts *TranscribeService) runSession(
	ctx context.Context,
	sess *TranscribeSession,
	frames <-chan ffmpeg.Frame,
	recorderErrs <-chan error,
) {
	defer func() {
		ts.mu.Lock()
		delete(ts.sessions, sess.ID)
		ts.mu.Unlock()
		ts.emitState(sess.ID, "stopped", "Recording stopped")
		close(sess.Done)
	}()

	jobs := make(chan Job, 4)

	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		ts.runWorker(sess.ID, jobs)
	}()
	defer func() {
		close(jobs)
		workers.Wait()
	}()

	audioChunker := chunker.NewAudioChunker()

	var chunkID int64
	for {
		select {
		case frame, ok := <-frames:
			if !ok {
				ts.enqueueChunks(context.Background(), jobs, audioChunker.Flush(), &chunkID)
				return
			}
			ts.enqueueChunks(ctx, jobs, audioChunker.AddFrame(frame), &chunkID)

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
			ts.enqueueChunks(context.Background(), jobs, audioChunker.Flush(), &chunkID)
			return
		}
	}
}

func (ts *TranscribeService) enqueueChunks(
	ctx context.Context,
	jobs chan<- Job,
	chunks []chunker.AudioChunk,
	chunkID *int64,
) {
	for _, chunk := range chunks {
		*chunkID = *chunkID + 1
		job := Job{ID: *chunkID, Chunk: chunk}
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
