package automatic1111

import (
	"bytes"
	"encoding/json"
	"io"
)

type Client struct {
	*Renderer
}

func New(options ...Option) (*Client, error) {
	r, err := NewRenderer(options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Renderer: r,
	}, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
