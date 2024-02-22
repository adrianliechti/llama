package ollama

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
	var err error

	c := &Client{}

	c.Embedder, err = NewEmbedder(url, options...)

	if err != nil {
		return nil, err
	}

	c.Completer, err = NewCompleter(url, options...)

	if err != nil {
		return nil, err
	}

	return c, nil
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
