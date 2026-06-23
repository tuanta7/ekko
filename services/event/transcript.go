package event

type TranscriptEvent struct {
	SessionID string `json:"sessionID"`
	ChunkID   int64  `json:"chunkID"`
	Text      string `json:"text"`
	Final     bool   `json:"final"`
	StartMs   int64  `json:"startMs"`
	EndMs     int64  `json:"endMs"`
}
