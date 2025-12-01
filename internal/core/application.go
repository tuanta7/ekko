package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tuanta7/transcript/internal/audio"
	"github.com/tuanta7/transcript/internal/transcriber"
	"github.com/tuanta7/transcript/pkg/queue"
	"github.com/tuanta7/transcript/pkg/x"
)

type TranscriptionChunk struct {
	Timestamp int64  `json:"timestamp"`
	Text      string `json:"text"`
	Error     error  `json:"error,omitempty"`
}

type Application struct {
	wg            sync.WaitGroup
	mu            sync.Mutex
	isRunning     atomic.Bool
	transcription *sync.Map

	ctx    context.Context
	cancel context.CancelFunc

	queue    *queue.RecordQueue
	counter  atomic.Uint32
	recorder *audio.Recorder
	trClient transcriber.Client
}

func NewApplication(recorder *audio.Recorder, client transcriber.Client) *Application {
	return &Application{
		recorder: recorder,
		trClient: client,
	}
}

func (a *Application) Start(chunkDuration time.Duration) (<-chan string, error) {
	// use mutex so initialization of ctx/cancel/queue/etc is serialized
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning.Load() {
		return nil, errors.New("session already running")
	}

	if err := a.initSession(); err != nil {
		return nil, err
	}

	a.isRunning.Store(true)

	stream := make(chan string, 100)
	_ = os.Mkdir(".tmp", 0755)

	a.wg = sync.WaitGroup{}
	a.wg.Add(2)

	go func() {
		defer func() {
			a.wg.Done()
			// time.Sleep(100 * time.Second) // for debugging
			close(stream) // Signal consumers when done
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "transcribe panic recovered: %v\n", r)
			}
		}()

		if err := a.transcribe(stream); err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		defer func() {
			a.wg.Done()
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "record panic recovered: %v\n", r)
			}
		}()

		if err := a.record(chunkDuration); err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		// when both workers finish, mark the session as not running
		a.wg.Wait()
		a.isRunning.Store(false)
	}()

	return stream, nil
}

func (a *Application) initSession() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.transcription = &sync.Map{}
	a.queue = queue.NewRecordQueue()
	a.counter.Store(0)

	err := a.trClient.ResetContext(context.TODO())
	if err != nil {
		// cleanup on failure
		if a.cancel != nil {
			a.cancel()
		}
		return fmt.Errorf("failed to reset transcriber context: %w", err)
	}

	return nil
}

func (a *Application) Stop() (string, error) {
	// lock to safely read/act on ctx/cancel and serialization with Start
	a.mu.Lock()
	if !a.isRunning.Load() {
		a.mu.Unlock()
		return "", errors.New("no active session")
	}

	if a.cancel != nil {
		a.cancel()
	}
	a.mu.Unlock()

	ch := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(ch)
	}()

	select {
	case <-ch: // workers finished
	case <-time.After(5 * time.Second):
	}

	filename, err := a.save()
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (a *Application) save() (string, error) {
	data, err := x.MarshalIndentSyncMap(a.transcription)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	filename := fmt.Sprintf("transcript-%s.json", time.Now().Format("20060102-150405"))
	if err = os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	_ = os.RemoveAll(".tmp")
	return filename, nil
}
