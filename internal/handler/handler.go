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

func (h *Handler) Close() {
	h.counter.Store(0)
	err := h.scheduler.Close()
	if err != nil {
		h.logger.Error("failed to close scheduler", zap.Error(err))
	}
}

func (h *Handler) ListSources(ctx context.Context) ([]string, error) {
	return h.recorder.ListSources(ctx)
}

func (h *Handler) StartRecord(ctx context.Context, chunkDuration time.Duration, source ...string) error {
	if len(source) == 0 {
		defaultSources, err := h.recorder.ListSources(ctx)
		if err != nil {
			h.logger.Error("failed to list sources", zap.Error(err))
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
				h.logger.Error("recording failed", zap.Error(err))
				return err
			}

			err = h.scheduler.Enqueue(ctx, &Job{
				SeqNumber: h.counter.Add(1),
				RawData:   data,
			})
			if err != nil {
				h.logger.Error("failed to enqueue job", zap.Error(err))
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
				h.logger.Error("transcription failed", zap.Error(err))
				return err
			}
			select {
			case h.out <- TranscriptResult{
				SeqNumber: job.SeqNumber,
				Text:      text,
			}:
			case <-ctx.Done():
				h.logger.Warn("context cancelled, dropping transcription result", zap.Uint32("seq", job.SeqNumber))
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
	var nextSeq uint32 = 1
	buffer := make(map[uint32]string)

	timer := time.NewTimer(flushTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("context cancelled, flushing remaining results")
			nextSeq = h.flushAll(buffer, nextSeq, output)
			return
		case <-timer.C: // Timeout: flush what we can, skip gaps if necessary
			if len(buffer) > 0 {
				h.logger.Warn("flush timeout, skipping gaps",
					zap.Uint32("expected", nextSeq),
					zap.Int("buffered", len(buffer)),
				)
				nextSeq = h.flushAll(buffer, nextSeq, output)
			}
			timer.Reset(flushTimeout)
		case result, ok := <-h.out:
			if !ok {
				h.logger.Info("transcript output channel closed, flushing remaining results")
				nextSeq = h.flushAll(buffer, nextSeq, output)
				return
			}

			buffer[result.SeqNumber] = result.Text
			nextSeq = h.flushNext(buffer, nextSeq, output)

			// Reset timer
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			timer.Reset(flushTimeout)
			h.logger.Debug("transcript result received", zap.Uint32("seq", result.SeqNumber))
		}
	}
}

// flushNext outputs buffered results in order from the given sequence number
func (h *Handler) flushNext(buffer map[uint32]string, nextSeq uint32, output func(text string)) uint32 {
	for {
		text, ok := buffer[nextSeq]
		if !ok {
			break
		}

		output(text)
		delete(buffer, nextSeq)
		nextSeq++
	}

	return nextSeq
}

// flushAll outputs buffered results in order, starting from the minimum sequence number
func (h *Handler) flushAll(buffer map[uint32]string, nextSeq uint32, output func(text string)) uint32 {
	if len(buffer) == 0 {
		h.logger.Warn("no buffered results to flush")
		return 1
	}

	// Find minimum seq in the buffer to continue from
	minSeq := nextSeq
	for seq := range buffer {
		if seq < minSeq || minSeq == nextSeq {
			minSeq = seq
		}
	}

	return h.flushNext(buffer, minSeq, output)
}
