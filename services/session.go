package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
	"github.com/tuanta7/ekko/services/chunker"
)

const (
	DefaultJobBufferSize = 4
)

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
	<-t.Done // Wait for the session to finish
}

func (t *TranscribeService) runSession(
	ctx context.Context,
	session *TranscribeSession,
	frames <-chan ffmpeg.Frame,
	recorderErrs <-chan error,
) {
	defer func() {
		t.mu.Lock()
		delete(t.sessions, session.ID)
		t.mu.Unlock()
		// Notify listeners that recording and transcription have stopped for this session.
		t.emitState(session.ID, EventRecordingStopped, "Recording stopped")
		close(session.Done)
	}()

	jobQueue := make(chan Job, DefaultJobBufferSize)

	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for job := range jobQueue {
			t.process(session.ID, job)
		}
	}()
	defer func() {
		close(jobQueue)
		workers.Wait()
	}()

	audioChunker := chunker.NewAudioChunker()

	var chunkID int64
	for {
		select {
		case frame, ok := <-frames:
			if !ok { // If the channel is close, flush the remaining chunks
				t.enqueueJob(context.Background(), jobQueue, audioChunker.Flush(), &chunkID)
				return
			}

			// Add the frame to the chunker and enqueue any new chunks
			t.enqueueJob(ctx, jobQueue, audioChunker.AddFrame(frame), &chunkID)

		case err, ok := <-recorderErrs:
			if !ok {
				recorderErrs = nil
				continue
			}
			if ok && err != nil {
				t.emitError(session.ID, err)
				return
			}

		case <-ctx.Done():
			t.enqueueJob(context.Background(), jobQueue, audioChunker.Flush(), &chunkID)
			return
		}
	}
}

func (t *TranscribeService) enqueueJob(
	ctx context.Context,
	queue chan<- Job,
	chunks []chunker.AudioChunk,
	chunkID *int64, // use pointer to persist increment across repeated calls
) {
	for _, chunk := range chunks {
		*chunkID = *chunkID + 1
		job := Job{
			ID:    *chunkID,
			Chunk: chunk,
		}

		if chunk.Final {
			select {
			case queue <- job:
			case <-ctx.Done():
				return
			}
			continue
		}

		select {
		case queue <- job:
		default:
		}
	}
}
