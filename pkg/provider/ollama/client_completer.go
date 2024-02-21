package ollama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func NewCompleter(url string, options ...Option) (*openai.Completer, error) {
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

	return openai.NewCompleter(c.options...)
}
