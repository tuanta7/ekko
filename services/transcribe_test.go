package services

import (
	"context"
	"testing"
	"time"
)

func TestStopCancelsWithoutWaitingForSessionDrain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	session := &TranscribeSession{
		ID:     "session-1",
		Cancel: cancel,
		Done:   make(chan struct{}),
	}
	service := &TranscribeService{
		sessions: map[string]*TranscribeSession{session.ID: session},
	}

	started := time.Now()
	if err := service.Stop(session.ID); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Stop() took %s; expected cancellation to return immediately", elapsed)
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("Stop() did not cancel the session context")
	}
}
