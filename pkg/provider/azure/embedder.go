package azure

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Embedder = openai.Embedder

func NewEmbedder(model string, options ...Option) (*Embedder, error) {
	c := &Config{
		options: []openai.Option{
			openai.WithURL("https://models.inference.ai.azure.com"),
		},
	}

	for _, option := range options {
		option(c)
	}

	return openai.NewEmbedder(model, c.options...)
}
