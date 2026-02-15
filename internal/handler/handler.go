package handler

import (
	"context"
	"fmt"
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
	out       chan string
}

func NewHandler(recorder *ffmpeg.Recorder, scriber *whisper.Client, logger *logger.Logger) *Handler {
	return &Handler{
		recorder: recorder,
		scriber:  scriber,
		logger:   logger,
		out:      make(chan string, 100),
	}
}

func (h *Handler) StartRecord(ctx context.Context, chunkDuration time.Duration, source ...string) error {
	errChan := make(chan error)
	defer close(errChan)

	if len(source) == 0 {
		defaultSources, err := h.recorder.ListSources(ctx)
		if err != nil {
			return err
		}
		source = defaultSources
	} else {
		h.recorder.SetSource(source[0])
	}

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

func (h *Handler) StartTranscribe(ctx context.Context) {
	h.scheduler.ProcessJobs(ctx, func(ctx context.Context, job *Job) error {
		return h.scriber.Transcribe(ctx, job.RawData, h.out)
	})
}

func (h *Handler) CollectResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-h.out:
			fmt.Println(result)
		}
	}
}

func (h *Handler) ListSources(ctx context.Context) ([]string, error) {
	return h.recorder.ListSources(ctx)
}

func (h *Handler) Stop() error {
	h.counter.Store(0)
	return nil
}
