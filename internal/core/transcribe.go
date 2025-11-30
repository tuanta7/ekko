package core

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/tuanta7/transcript/pkg/queue"
)

// transcribe continuously dequeues and transcribes audio files
func (a *Application) transcribe(stream chan string) error {
	for {
		msg, err := a.queue.Dequeue(a.ctx)
		if err != nil {
			if errors.Is(err, queue.ErrQueueClosed) {
				return nil
			}
			return err
		}

		reader, err := a.trClient.Transcribe(a.ctx, msg.FileName)
		if err != nil {
			if a.ctx.Err() != nil {
				return a.ctx.Err()
			}
			return fmt.Errorf("failed to transcribe audio: %w", err)
		}

		fullText := ""
		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			chunk := scanner.Text()
			fullText += chunk

			select {
			case stream <- chunk + "\n":
			case <-a.ctx.Done():
				_ = reader.Close()
				return a.ctx.Err()
			}
		}

		transcriptionErr := scanner.Err()
		if transcriptionErr != nil {
			errMsg := fmt.Sprintf("[Error] Failed to read transcript: %s\n", transcriptionErr.Error())

			select {
			case stream <- errMsg:
			case <-a.ctx.Done():
				_ = reader.Close()
				return a.ctx.Err()
			}
		}

		_ = reader.Close()
		_ = os.Remove(msg.FileName)

		a.transcription.Store(time.Now().Format(time.RFC822), TranscriptionChunk{
			Timestamp: msg.Timestamp.Unix(),
			Text:      fullText,
			Error:     transcriptionErr,
		})
	}
}
