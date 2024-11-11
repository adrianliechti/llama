package google

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(model string, options ...Option) (*Completer, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/"

	c := &Config{}

	for _, option := range options {
		option(c)
	}

	return openai.NewCompleter(url, model, c.options...)
}
