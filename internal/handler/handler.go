package handler

import (
	"context"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/tuanta7/ekko/internal/ffmpeg"
	"github.com/tuanta7/ekko/internal/whisper"
	"github.com/tuanta7/ekko/pkg/logger"
)

const (
	flushTimeout = 5 * time.Second
)

type TranscriptResult struct {
	SeqNumber uint32
	Text      string
}

type Handler struct {
	counter   atomic.Uint32
	recorder  *ffmpeg.Recorder
	scriber   *whisper.Client
	scheduler *Scheduler
	logger    *logger.Logger
	out       chan TranscriptResult
}

func NewHandler(recorder *ffmpeg.Recorder, scriber *whisper.Client, logger *logger.Logger) *Handler {
	return &Handler{
		recorder:  recorder,
		scriber:   scriber,
		scheduler: New(100), // Initialize scheduler with buffer size
		logger:    logger,
		out:       make(chan TranscriptResult, 100),
	}
}

func (h *Handler) StartRecord(ctx context.Context, chunkDuration time.Duration, source ...string) error {
	if len(source) == 0 {
		defaultSources, err := h.recorder.ListSources(ctx)
		if err != nil {
			return err
		}
		h.recorder.SetSource(defaultSources[0])
	} else {
		h.recorder.SetSource(source[0])
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			data, err := h.recorder.Record(ctx, chunkDuration)
			if err != nil {
				return err
			}

			err = h.scheduler.Enqueue(ctx, &Job{
				SeqNumber: h.counter.Add(1),
				RawData:   data,
			})
			if err != nil {
				return err
			}
		}
	}
}

func (h *Handler) StartTranscribe(ctx context.Context) {
	h.scheduler.ProcessJobs(ctx,
		func(ctx context.Context, job *Job) error {
			text, err := h.scriber.Transcribe(ctx, job.RawData)
			if err != nil {
				return err
			}
			select {
			case h.out <- TranscriptResult{
				SeqNumber: job.SeqNumber,
				Text:      text,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		},
		func(job *Job, err error) {
			h.logger.Error("transcription failed",
				zap.Uint32("seq", job.SeqNumber),
				zap.Error(err),
			)
		},
	)
}

// CollectResults collects transcription results and outputs them in order.
// It buffers out-of-order results and flushes them after a timeout to prevent
// indefinite waiting for missing sequences.
func (h *Handler) CollectResults(ctx context.Context, output func(text string)) {
	timer := time.NewTimer(flushTimeout)
	defer timer.Stop()

	var nextSeq uint32 = 1
	buffer := make(map[uint32]string)

	flush := func() { // Output buffered results in order
		for {
			if text, ok := buffer[nextSeq]; ok {
				output(text)
				delete(buffer, nextSeq)
				nextSeq++
			} else {
				break
			}
		}
	}

	flushAll := func() {
		if len(buffer) == 0 {
			return
		}

		// Find minimum seq in the buffer to continue from
		minSeq := nextSeq
		for seq := range buffer {
			if seq < minSeq || minSeq == nextSeq {
				minSeq = seq
			}
		}
		nextSeq = minSeq
		flush()
	}

	for {
		select {
		case <-ctx.Done():
			flushAll()
			return

		case result, ok := <-h.out:
			if !ok {
				flushAll()
				return
			}

			buffer[result.SeqNumber] = result.Text
			flush()

			// Reset timer
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(flushTimeout)

		case <-timer.C: // Timeout: flush what we can, skip gaps if necessary
			if len(buffer) > 0 {
				h.logger.Warn("flush timeout, skipping gaps",
					zap.Uint32("expected", nextSeq),
					zap.Int("buffered", len(buffer)),
				)
				flushAll()
			}
			timer.Reset(flushTimeout)
		}
	}
}

func (h *Handler) ListSources(ctx context.Context) ([]string, error) {
	return h.recorder.ListSources(ctx)
}

func (h *Handler) Close() {
	h.counter.Store(0)
	err := h.scheduler.Close()
	if err != nil {
		h.logger.Error("failed to close scheduler", zap.Error(err))
	}
}
