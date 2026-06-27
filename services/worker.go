package services

import (
	"github.com/tuanta7/ekko/services/adapter/whisper"
	"github.com/tuanta7/ekko/services/chunker"
)

type Job struct {
	ID    int64
	Chunk chunker.AudioChunk
}

// process transcribes one queued audio chunk and emits its transcript and session state events.
func (t *TranscribeService) process(sessionID string, job Job) {
	// Notify listeners that transcription is in progress for this session.
	t.emitState(sessionID, EventTranscribing, "")

	segments, err := t.scriber.Transcribe(job.Chunk.Samples, whisper.TranscribeOptions{
		TokenTimestamps: job.Chunk.Final,
	})
	if err != nil {
		t.emitError(sessionID, err)
		return
	}

	text := whisper.CombineSegments(segments)
	if text == "" {
		return
	}

	t.emitTranscript(TranscriptEvent{
		SessionID: sessionID,
		ChunkID:   job.ID,
		Text:      text,
		Final:     job.Chunk.Final,
		StartMs:   job.Chunk.Start.Milliseconds(),
		EndMs:     job.Chunk.End.Milliseconds(),
	})

	if job.Chunk.Final {
		// Notify listeners that the session has returned to recording after the final chunk.
		t.emitState(sessionID, EventRecording, "")
	}
}
