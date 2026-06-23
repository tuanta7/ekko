package session

import (
	"context"
	"fmt"
	"time"

	"github.com/tuanta7/ekko/services/adapter/ffmpeg"
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
	<-t.Done
}

type StartOptions struct {
	Source    string `json:"source"`
	ModelName string `json:"modelName"`
	Language  string `json:"language"`
	Threads   uint   `json:"threads"`
	Translate bool   `json:"translate"`
}

func (t *TranscribeSession) Run(
	ctx context.Context,
	frames <-chan ffmpeg.Frame,
	errCh <-chan error,
	options StartOptions,
) {

}
