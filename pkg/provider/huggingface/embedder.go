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
	if url == "" {
		return nil, errors.New("invalid url")
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

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

func (e *Embedder) Embed(ctx context.Context, content string) (provider.Embeddings, error) {
	body := map[string]any{
		"inputs": strings.TrimSpace(content),
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", e.url, jsonReader(body))
	req.Header.Set("Authorization", "Bearer "+e.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
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

	return nil, errors.New("unable to embed input")
}
