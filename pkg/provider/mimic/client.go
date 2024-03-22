package mimic

import (
	"bytes"
	"encoding/json"
	"io"

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

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
