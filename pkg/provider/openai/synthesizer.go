package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var _ provider.Synthesizer = (*Synthesizer)(nil)

type Synthesizer struct {
	*Config
	client *openai.Client
}

func NewSynthesizer(options ...Option) (*Synthesizer, error) {
	cfg := &Config{
		model: string(openai.TTSModel1),
	}

	for _, option := range options {
		option(cfg)
	}

	return &Synthesizer{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (s *Synthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	if options == nil {
		options = new(provider.SynthesizeOptions)
	}

	req := openai.CreateSpeechRequest{
		Input: content,

		Model: openai.SpeechModel(s.model),
		Voice: openai.VoiceAlloy,

		ResponseFormat: openai.SpeechResponseFormatWav,
	}

	result, err := s.client.CreateSpeech(ctx, req)

	if err != nil {
		convertError(err)
	}

	id := uuid.New().String()

	return &provider.Synthesis{
		ID: id,

		Name:    id + ".wav",
		Content: result,
	}, nil
}
