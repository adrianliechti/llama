package coqui

import (
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Synthesizer
}

func New(url string, options ...Option) (*Client, error) {
	s, err := NewSynthesizer(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Synthesizer: s,
	}, nil
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
