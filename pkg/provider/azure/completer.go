package azure

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(url, model string, options ...Option) (*Completer, error) {
	if url == "" {
		url = "https://models.inference.ai.azure.com"
	}

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	return openai.NewCompleter(url, model, cfg.options...)
}
