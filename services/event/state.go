package event

const (
	StateEventName = "transcribe:state"
)

type StateEvent struct {
	SessionID string `json:"sessionID"`
	State     string `json:"state"`
	Message   string `json:"message"`
}
