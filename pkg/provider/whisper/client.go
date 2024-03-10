package whisper

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Transcriber
}

func New(url string, options ...Option) (*Client, error) {
	t, err := NewTranscriber(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Transcriber: t,
	}, nil
}
