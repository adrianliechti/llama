package openai

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Embedder
	provider.Completer
	provider.Transcriber
}

func New(options ...Option) (*Client, error) {
	var err error

	c := &Client{}

	c.Embedder, err = NewEmbedder(options...)

	if err != nil {
		return nil, err
	}

	c.Completer, err = NewCompleter(options...)

	if err != nil {
		return nil, err
	}

	c.Transcriber, err = NewTranscriber(options...)

	if err != nil {
		return nil, err
	}

	return c, nil
}
