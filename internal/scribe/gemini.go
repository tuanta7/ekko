package scribe

import (
	"context"
	"fmt"
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
		model:  "gemini-2.0-flash-lite",
	}, nil
}

func (c *GeminiClient) ResetContext() error {
	return nil
}

func (c *GeminiClient) Close() error {
	return nil
}

func (c *GeminiClient) Transcribe(ctx context.Context, audioPath string, out chan string) error {
	contents, err := c.newContentsFromAudio(audioPath)
	if err != nil {
		return err
	}

	if c == nil || c.client == nil {
		return fmt.Errorf("gemini client not initialized")
	}

	temperature := float32(0.5)
	stream := c.client.Models.GenerateContentStream(ctx, c.model, contents, &genai.GenerateContentConfig{
		Temperature: &temperature,
	})

	if stream == nil {
		return fmt.Errorf("generate content stream returned nil")
	}

	for chunk, chunkErr := range stream {
		if chunkErr != nil {
			c.writeToStream(ctx, out, normalizeError(chunkErr))
			continue
		}
		c.writeToStream(ctx, out, chunk)
	}

	return nil
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
		genai.NewPartFromText(InitialPrompts),
		audioPart,
	}

	return []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}, nil
}

func (c *GeminiClient) writeToStream(ctx context.Context, out chan string, chunk *genai.GenerateContentResponse) {
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

				out <- part.Text
			}
		}
	}
}

func normalizeError(chunkErr error) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						genai.NewPartFromText(fmt.Sprintf("Error while process audio: %v", chunkErr)),
					},
				},
			},
		},
	}
}
