package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

func toFloat32s(v []float64) []float32 {
	result := make([]float32, len(v))

	for i, x := range v {
		result[i] = float32(x)
	}

	return result
}
