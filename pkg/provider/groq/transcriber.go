package groq

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Transcriber = openai.Transcriber

func NewTranscriber(url, model string, options ...Option) (*Transcriber, error) {
	if url == "" {
		url = "https://api.groq.com/openai/v1"
	}

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewTranscriber(url, model, c.options...)
}
