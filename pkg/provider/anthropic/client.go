package anthropic

import (
	"bytes"
	"encoding/json"
	"io"
)

type Client struct {
	*Completer
}

func New(options ...Option) (*Client, error) {
	c, err := NewCompleter(options...)

	if err != nil {
		return nil, err
	}

	return &Client{
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
