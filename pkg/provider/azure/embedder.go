package azure

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Embedder = openai.Embedder

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "https://models.inference.ai.azure.com"
	}

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	return openai.NewEmbedder(url, model, cfg.options...)
}
