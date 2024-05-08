package deepl

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

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

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
