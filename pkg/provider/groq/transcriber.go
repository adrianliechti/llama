package groq

import (
	"github.com/adrianliechti/wingman/pkg/provider/openai"
)

type Transcriber = openai.Transcriber

func NewTranscriber(url, model string, options ...Option) (*Transcriber, error) {
	if url == "" {
		url = "https://api.groq.com/openai/v1"
	}

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	return openai.NewTranscriber(url, model, cfg.options...)
}
