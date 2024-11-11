package google

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Embedder = openai.Embedder

func NewEmbedder(model string, options ...Option) (*Embedder, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/"

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewEmbedder(url, model, c.options...)
}
