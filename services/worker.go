package services

import (
	"github.com/tuanta7/ekko/services/adapter/whisper"
)

func (ts *TranscribeService) runWorker(
	sessionID string,
	jobs <-chan Job,
) {
	for job := range jobs {
		ts.process(sessionID, job)
	}
}

func (ts *TranscribeService) process(sessionID string, job Job) {
	ts.emitState(sessionID, "transcribing", "")

	segments, err := ts.scriber.Transcribe(job.Chunk.Samples, whisper.TranscribeOptions{
		TokenTimestamps: job.Chunk.Final,
	})
	if err != nil {
		ts.emitError(sessionID, err)
		return
	}

	text := whisper.CombineSegments(segments)
	if text == "" {
		return
	}

	// New transcript to be displayed
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
