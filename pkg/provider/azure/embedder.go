package azure

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Embedder = openai.Embedder

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "https://models.inference.ai.azure.com"
	}

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewEmbedder(url, model, c.options...)
}
