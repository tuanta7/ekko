package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tuanta7/ekko/internal/audio"
	"github.com/tuanta7/ekko/internal/scribe"
	"github.com/tuanta7/ekko/pkg/logger"
	"github.com/tuanta7/ekko/pkg/queue"
)

const (
	dataFileNameKey = "fileName"
)

type TranscriptionChunk struct {
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
	Error     error  `json:"error,omitempty"`
}

type Handler struct {
	wg          sync.WaitGroup
	mu          sync.Mutex
	isRunning   bool
	sessionFile *os.File
	logger      *logger.FileLogger
	queue       *queue.RecordQueue
	counter     atomic.Uint32
	recorder    audio.Recorder
	scribe      scribe.Scriber
}

func NewHandler(recorder audio.Recorder, scribe scribe.Scriber, logger *logger.FileLogger) *Handler {
	return &Handler{
		recorder: recorder,
		scribe:   scribe,
		logger:   logger,
	}
}

func (h *Handler) Start(ctx context.Context, chunkDuration time.Duration) (<-chan string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.isRunning {
		return nil, errors.New("session already running")
	}

	if err := os.RemoveAll(".tmp"); err != nil {
		return nil, err
	}

	if err := os.Mkdir(".tmp", 0755); err != nil {
		return nil, err
	}

	h.wg = sync.WaitGroup{}
	h.queue = queue.NewRecordQueue()
	h.counter.Store(0)

	filename := fmt.Sprintf("%s.txt", time.Now().Format(time.RFC3339))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create session file: %w", err)
	}
	h.sessionFile = f

	if err = h.scribe.ResetContext(); err != nil {
		return nil, fmt.Errorf("failed to reset transcriber context: %w", err)
	}

	h.isRunning = true

	stream := make(chan string, 100)
	h.wg.Add(2)

	go func() {
		defer func() {
			h.wg.Done()
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "record panic recovered: %v\n", r)
			}
		}()

		if err := h.record(ctx, chunkDuration); err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		defer func() {
			h.wg.Done()
			close(stream) // signal consumers when done
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "transcribe panic recovered: %v\n", r)
			}
		}()

		if err := h.transcribe(ctx, stream); err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		h.wg.Wait()
		h.isRunning = false
	}()

	return stream, nil
}

func (h *Handler) SaveChunk(chunk string) error {
	text := fmt.Sprintf("%s: %s\n", time.Now().Format(time.TimeOnly), chunk)
	h.mu.Lock()
	_, err := h.sessionFile.Write([]byte(text))
	h.mu.Unlock()
	return err
}

func (h *Handler) Stop() error {
	h.mu.Lock()
	if !h.isRunning {
		h.mu.Unlock()
		return errors.New("no active session")
	}

	if err := os.RemoveAll(".tmp"); err != nil {
		return err
	}

	h.mu.Unlock()

	ch := make(chan struct{})
	go func() {
		h.wg.Wait() // workers finished
		close(ch)
	}()

	select {
	case <-ch: // channel closed
	case <-time.After(5 * time.Second):
	}

	if h.sessionFile != nil {
		if err := h.sessionFile.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		h.sessionFile = nil
	}

	return nil
}
