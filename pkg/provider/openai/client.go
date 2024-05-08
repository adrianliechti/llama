package openai

import (
	"errors"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	*Embedder
	*Completer
	*Synthesizer
	*Transcriber
	*Renderer
}

func New(options ...Option) (*Client, error) {
	e, err := NewEmbedder(options...)

	if err != nil {
		return nil, err
	}

	c, err := NewCompleter(options...)

	if err != nil {
		return nil, err
	}

	s, err := NewSynthesizer(options...)

	if err != nil {
		return nil, err
	}

	t, err := NewTranscriber(options...)

	if err != nil {
		return nil, err
	}

	r, err := NewRenderer(options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Embedder:    e,
		Completer:   c,
		Synthesizer: s,
		Transcriber: t,
		Renderer:    r,
	}, nil
}

func convertError(err error) error {
	var oaierr *openai.APIError

	if errors.As(err, &oaierr) {
		return errors.New(oaierr.Message)
	}

	return err
}
