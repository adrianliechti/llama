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

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewEmbedder(url+"/v1", model, c.options...)
}
