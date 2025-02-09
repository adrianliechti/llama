package mistral

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(model string, options ...Option) (*Completer, error) {
	url := "https://api.mistral.ai/v1/"

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	return openai.NewCompleter(url, model, cfg.options...)
}
