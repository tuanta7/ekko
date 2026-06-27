package services

const (
	EventTranscribing     = "transcribing"
	EventRecording        = "recording"
	EventRecordingStopped = "stopped"

	EventState   = "transcribe:state"
	EventPartial = "transcribe:partial"
	EventFinal   = "transcribe:final" // final chunk
	EventError   = "transcribe:error"
)

type StateEvent struct {
	SessionID string `json:"sessionID"`
	State     string `json:"state"`
	Message   string `json:"message"`
}

type TranscriptEvent struct {
	SessionID string `json:"sessionID"`
	ChunkID   int64  `json:"chunkID"`
	Text      string `json:"text"`
	Final     bool   `json:"final"`
	StartMs   int64  `json:"startMs"`
	EndMs     int64  `json:"endMs"`
}

type ErrorEvent struct {
	SessionID string `json:"sessionID"`
	Message   string `json:"message"`
}

func (t *TranscribeService) emit(name string, data any) {
	t.mu.Lock()
	app := t.app // Lock briefly to safely read the shared app
	t.mu.Unlock()

	if app != nil {
		app.Event.Emit(name, data)
	}
}

// emitState publishes the current transcription session state to application listeners.
func (t *TranscribeService) emitState(sessionID string, state string, message string) {
	t.emit(EventState, StateEvent{
		SessionID: sessionID,
		State:     state,
		Message:   message,
	})
}

func (t *TranscribeService) emitTranscript(event TranscriptEvent) {
	if event.Final {
		t.emit(EventFinal, event)
		return
	}

	t.emit(EventPartial, event)
}

func (t *TranscribeService) emitError(sessionID string, err error) {
	if err == nil {
		return
	}

	t.emit(EventError, ErrorEvent{
		SessionID: sessionID,
		Message:   err.Error(),
	})
}
