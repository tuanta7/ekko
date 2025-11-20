package main

import (
	"context"
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

func (c *GeminiClient) NewContentsFromAudio(ctx context.Context, audioPath string) ([]*genai.Content, error) {
	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, err
	}

	audioPart := &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: "audio/wav",
			Data:     audioBytes,
		},
	}

	parts := []*genai.Part{
		genai.NewPartFromText("Generate a transcript of the speech."),
		audioPart,
	}
	return []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}, nil
}

func (c *GeminiClient) Transcribe(ctx context.Context, contents []*genai.Content) (string, error) {
	temperature := float32(0.2)
	content, err := c.client.Models.GenerateContent(ctx, c.model, contents, &genai.GenerateContentConfig{
		Temperature: &temperature,
	})
	if err != nil {
		return "", err
	}

	return content.Text(), nil
}
