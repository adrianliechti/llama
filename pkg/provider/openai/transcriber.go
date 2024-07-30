package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var _ provider.Transcriber = (*Transcriber)(nil)

type Transcriber struct {
	*Config
	client *openai.Client
}

func NewTranscriber(options ...Option) (*Transcriber, error) {
	cfg := &Config{
		model: openai.Whisper1,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Transcriber{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (c *Transcriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if options == nil {
		options = new(provider.TranscribeOptions)
	}

	id := uuid.NewString()

	req := openai.AudioRequest{
		Model: c.model,

		Language: options.Language,

		Reader:   input.Content,
		FilePath: input.Name,
	}

	transcription, err := c.client.CreateTranscription(ctx, req)

	if err != nil {
		convertError(err)
	}

	result := provider.Transcription{
		ID: id,

		Content: transcription.Text,
	}

	return &result, nil
}
