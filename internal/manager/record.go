package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/tuanta7/ekko/internal/queue"
)

// record continuously records audio and enqueues for transcription
func (h *Handler) record(ctx context.Context, duration time.Duration) error {
	defer h.queue.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			currentCount := h.counter.Add(1)
			fileName := fmt.Sprintf(".tmp/audio-%d.wav", currentCount)

			if err := h.recorder.Record(ctx, duration, fileName); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("recording failed: %w", err)
			}

			if err := h.queue.Enqueue(ctx, &queue.Message{
				Timestamp: time.Now(),
				Data: map[string]string{
					dataFileNameKey: fileName,
				},
			}); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
			}
		}
	}
}
