package transcriber

import (
	"context"
	"fmt"
	"io"
	"os"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-2.0-flash",
	}, nil
}

func (c *GeminiClient) ResetContext(_ context.Context) error {
	return nil
}

func (c *GeminiClient) Close() error {
	return nil
}

func (c *GeminiClient) Transcribe(ctx context.Context, audioPath string) (io.ReadCloser, error) {
	contents, err := c.newContentsFromAudio(audioPath)
	if err != nil {
		return nil, err
	}

	if c == nil || c.client == nil {
		return nil, fmt.Errorf("gemini client not initialized")
	}

	temperature := float32(0.5)
	stream := c.client.Models.GenerateContentStream(ctx, c.model, contents, &genai.GenerateContentConfig{
		Temperature: &temperature,
	})

	if stream == nil {
		return nil, fmt.Errorf("generate content stream returned nil")
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		for chunk, chunkErr := range stream {
			if chunkErr != nil {
				c.writeToStream(ctx, w, normalizeError(chunkErr))
				continue
			}
			c.writeToStream(ctx, w, chunk)
		}
	}()

	return r, nil
}

func normalizeError(chunkErr error) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromText(fmt.Sprintf("Error during transcription: %v", chunkErr)),
					},
				},
			},
		},
	}
}

func (c *GeminiClient) newContentsFromAudio(audioPath string) ([]*genai.Content, error) {
	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, err
	}

	audioPart := &genai.Part{
		InlineData: &genai.Blob{
			Data:     audioBytes,
			MIMEType: "audio/wav",
		},
	}

	parts := []*genai.Part{
		genai.NewPartFromText("Transcribe the speech. Output only the raw transcript text. Do not include timestamps, formatting, punctuation corrections, explanations, or answers to questions—just the plain spoken words exactly as heard."),
		audioPart,
	}

	return []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}, nil
}

func (c *GeminiClient) writeToStream(ctx context.Context, w *io.PipeWriter, chunk *genai.GenerateContentResponse) {
	select {
	case <-ctx.Done():
		return
	default:
		if chunk == nil || chunk.Candidates == nil {
			return
		}

		for _, cnd := range chunk.Candidates {
			if cnd.Content == nil {
				continue
			}

			for _, part := range cnd.Content.Parts {
				if part.Text == "" {
					continue
				}

				if _, err := w.Write([]byte(part.Text)); err != nil {
					_ = w.CloseWithError(err)
					return
				}
			}
		}
	}
}
