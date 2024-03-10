package deepl

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Completer
	provider.Translator
}

func New(url string, options ...Option) (*Client, error) {
	var err error

	t, err := NewTranslator(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Translator: t,
		Completer:  t,
	}, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
