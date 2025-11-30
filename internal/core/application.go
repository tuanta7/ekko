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

type Transcription struct {
	Timestamp time.Time `json:"timestamp"`
	FileName  string    `json:"filename"`
	Text      string    `json:"text"`
	Error     error     `json:"error,omitempty"`
}

type Session struct {
}

type Application struct {
	mu             sync.Mutex
	isRunning      bool
	transcriptions *sync.Map

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
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return nil, errors.New("session already running")
	}

	if err := a.reset(); err != nil {
		return nil, err
	}

	stream := make(chan string, 100)
	_ = os.Mkdir(".tmp", 0755)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer func() {
			wg.Done()
			close(stream) // Signal consumers when done
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "transcribe panic recovered: %v\n", r)
			}
		}()

		err := a.transcribe(stream)
		if err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		defer func() {
			wg.Done()
			if r := recover(); r != nil {
				_, _ = fmt.Fprintf(os.Stderr, "record panic recovered: %v\n", r)
			}
		}()

		err := a.record(chunkDuration)
		if err != nil && !errors.Is(err, context.Canceled) {
			stream <- err.Error()
		}
	}()

	go func() {
		wg.Wait()
		a.mu.Lock()
		a.isRunning = false
		a.mu.Unlock()
	}()

	return stream, nil
}

func (a *Application) reset() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.isRunning = true
	a.transcriptions = &sync.Map{}
	a.queue = queue.NewRecordQueue()
	a.counter.Store(0)

	err := a.trClient.ResetContext(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to reset transcriber context: %w", err)
	}

	return nil
}

func (a *Application) Stop() (string, error) {
	a.mu.Lock()
	if !a.isRunning {
		a.mu.Unlock()
		return "", errors.New("no active session")
	}

	if a.cancel != nil {
		a.cancel()
	}
	a.mu.Unlock()

	// graceful shutdown
	time.Sleep(time.Second)

	filename, err := a.save()
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (a *Application) save() (string, error) {
	a.mu.Lock()
	filename := fmt.Sprintf("transcript-%s.json", time.Now().Format("20060102-150405"))
	data, err := x.MarshalIndentSyncMap(a.transcriptions)
	a.mu.Unlock()

	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	if err = os.WriteFile(filename, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	_ = os.RemoveAll(".tmp")
	return filename, nil
}
