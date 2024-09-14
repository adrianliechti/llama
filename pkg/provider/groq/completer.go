package groq

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(model string, options ...Option) (*Completer, error) {
	url := "https://api.groq.com/openai/v1"

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewCompleter(url, model, c.options...)
}
