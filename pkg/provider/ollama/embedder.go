package ollama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Embedder = openai.Embedder

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	opts := []openai.Option{}

	if cfg.client != nil {
		opts = append(opts, openai.WithClient(cfg.client))
	}

	return openai.NewEmbedder(url+"/v1", model, opts...)
}
