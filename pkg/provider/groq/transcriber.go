package groq

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Transcriber = openai.Transcriber

func NewTranscriber(model string, options ...Option) (*Transcriber, error) {
	url := "https://api.groq.com/openai"

	c := &Config{
		options: []openai.Option{
			openai.WithURL(url + "/v1"),
		},
	}

	for _, option := range options {
		option(c)
	}

	return openai.NewTranscriber(model, c.options...)
}
