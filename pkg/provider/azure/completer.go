package azure

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(model string, options ...Option) (*Completer, error) {
	c := &Config{
		options: []openai.Option{
			openai.WithURL("https://models.inference.ai.azure.com"),
		},
	}

	for _, option := range options {
		option(c)
	}

	return openai.NewCompleter(model, c.options...)
}
