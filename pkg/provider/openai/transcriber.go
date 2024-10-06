package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
)

var _ provider.Transcriber = (*Transcriber)(nil)

type Transcriber struct {
	*Config
	transcriptions *openai.AudioTranscriptionService
}

func NewTranscriber(url, model string, options ...Option) (*Transcriber, error) {
	cfg := &Config{
		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Transcriber{
		Config:         cfg,
		transcriptions: openai.NewAudioTranscriptionService(cfg.Options()...),
	}, nil
}

func (t *Transcriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if options == nil {
		options = new(provider.TranscribeOptions)
	}

	id := uuid.NewString()

	transcription, err := t.transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		Model: openai.F(t.model),
		File:  openai.FileParam(input.Content, input.Name, ""),
	})

	if err != nil {
		return nil, convertError(err)
	}

	result := provider.Transcription{
		ID: id,

		Content: transcription.Text,
	}

	return &result, nil
}
