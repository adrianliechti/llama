package groq

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(options ...Option) (*Completer, error) {
	url := "https://api.groq.com/openai"

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
