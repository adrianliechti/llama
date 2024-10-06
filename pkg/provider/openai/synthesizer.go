package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/openai/openai-go"
)

var _ provider.Synthesizer = (*Synthesizer)(nil)

type Synthesizer struct {
	*Config
	speech *openai.AudioSpeechService
}

func NewSynthesizer(url, model string, options ...Option) (*Synthesizer, error) {
	cfg := &Config{
		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Synthesizer{
		Config: cfg,
		speech: openai.NewAudioSpeechService(cfg.Options()...),
	}, nil
}

func (s *Synthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	if options == nil {
		options = new(provider.SynthesizeOptions)
	}

	result, err := s.speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.F(s.model),
		Input: openai.F(content),

		Voice:          openai.F(openai.AudioSpeechNewParamsVoiceAlloy),
		ResponseFormat: openai.F(openai.AudioSpeechNewParamsResponseFormatWAV),
	})

	if err != nil {
		return nil, convertError(err)
	}

	id := uuid.NewString()

	return &provider.Synthesis{
		ID: id,

		Name:    id + ".wav",
		Content: result.Body,
	}, nil
}
