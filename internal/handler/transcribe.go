package handler

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tuanta7/ekko/pkg/queue"
)

func (h *Handler) transcribe(ctx context.Context, stream chan string) error {
	for {
		msg, err := h.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, queue.ErrQueueClosed) {
				return nil
			}
			return err
		}

		// synchronous processing
		if err = h.scribe.Transcribe(ctx, msg.Data[dataFileNameKey], stream); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("failed to transcribe audio: %w", err)
		}

		_ = os.Remove(msg.Data[dataFileNameKey])
	}
}
