package ollama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func NewEmbedder(url string, options ...Option) (*openai.Embedder, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	c := &Config{
		options: []openai.Option{
			openai.WithURL(url + "/v1"),
		},
	}

	for _, option := range options {
		option(c)
	}

	return openai.NewEmbedder(c.options...)
}
