package huggingface

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
}

func NewEmbedder(url string, options ...Option) (*Embedder, error) {
	cfg := &Config{
		url: url,

		token: "-",
		model: "tei",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) ([]float32, error) {
	body := map[string]any{
		"inputs": strings.TrimSpace(content),
	}

	resp, err := e.client.Post(e.url, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to encode input")
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var result1 []float32

	if err := json.Unmarshal(data, &result1); err == nil {
		return result1, nil
	}

	var result2 [][]float32

	if err := json.Unmarshal(data, &result2); err == nil && len(result2) > 0 {
		return result2[0], nil
	}

	return nil, errors.New("unable to decode output")
}
