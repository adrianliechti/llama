package llama

import (
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Client struct {
	*openai.Embedder
	*openai.Completer
}

func New(url string, options ...Option) (*Client, error) {
	var err error

	c := &Client{}

	c.Embedder, err = NewEmbedder(url, options...)

	if err != nil {
		return nil, err
	}

	c.Completer, err = NewCompleter(url, options...)

	if err != nil {
		return nil, err
	}

	return c, nil
}
