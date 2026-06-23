package event

type ErrorEvent struct {
	SessionID string `json:"sessionID"`
	Message   string `json:"message"`
}
