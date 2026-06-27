package services

const (
	EventTranscribing = "transcribing"
	EventState        = "transcribe:state"
	EventPartial      = "transcribe:partial"
	EventFinal        = "transcribe:final"
	EventError        = "transcribe:error"
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

func (ts *TranscribeService) emit(name string, data any) {
	ts.mu.Lock()
	app := ts.app // Lock briefly to safely read the shared app
	ts.mu.Unlock()

	if app != nil {
		app.Event.Emit(name, data)
	}
}

func (ts *TranscribeService) emitState(sessionID string, state string, message string) {
	ts.emit(EventState, StateEvent{
		SessionID: sessionID,
		State:     state,
		Message:   message,
	})
}

func (ts *TranscribeService) emitTranscript(event TranscriptEvent) {
	if event.Final {
		ts.emit(EventFinal, event)
		return
	}

	ts.emit(EventPartial, event)
}

func (ts *TranscribeService) emitError(sessionID string, err error) {
	if err == nil {
		return
	}

	ts.emit(EventError, ErrorEvent{
		SessionID: sessionID,
		Message:   err.Error(),
	})
}
