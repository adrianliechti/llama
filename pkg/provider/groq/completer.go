package groq

import (
	"github.com/adrianliechti/wingman/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(model string, options ...Option) (*Completer, error) {
	url := "https://api.groq.com/openai/v1"

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	return openai.NewCompleter(url, model, cfg.options...)
}
