package handler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/internal/whisper"
	"github.com/tuanta7/ekko/pkg/logger"
)

type Handler struct {
	wg        sync.WaitGroup
	mu        sync.Mutex
	counter   atomic.Uint32
	recorder  *ffmpeg.Recorder
	scriber   *whisper.Client
	scheduler *Scheduler
	logger    *logger.Logger
}

func NewHandler(recorder *ffmpeg.Recorder, scriber *whisper.Client, logger *logger.Logger) *Handler {
	return &Handler{
		recorder: recorder,
		scriber:  scriber,
		logger:   logger,
	}
}

func (h *Handler) StartRecord(ctx context.Context, chunkDuration time.Duration) error {
	errChan := make(chan error)
	defer close(errChan)

	go func() {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
			data, err := h.recorder.Record(ctx, chunkDuration)
			if err != nil {
				errChan <- err
				return
			}

			err = h.scheduler.Enqueue(ctx, &Job{
				SeqNumber: h.counter.Add(1),
				RawData:   data,
			})
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	err := <-errChan
	return err
}

func (h *Handler) StartTranscribe(ctx context.Context) {}

func (h *Handler) Stop() error {
	h.counter.Store(0)
	return nil
}
