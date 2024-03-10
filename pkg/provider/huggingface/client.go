package huggingface

import (
	"bytes"
	"encoding/json"
	"io"
)

type Client struct {
	*Embedder
	*Completer
}

func New(url string, options ...Option) (*Client, error) {
	e, err := NewEmbedder(url, options...)

	if err != nil {
		return nil, err
	}

	c, err := NewCompleter(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Embedder:  e,
		Completer: c,
	}, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
