package jina

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
}

func NewEmbedder(url string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "https://api.jina.ai"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{
		client: http.DefaultClient,

		url: url,

		model: "jina-clip-v1",
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	body := map[string]any{
		"input": []string{
			strings.TrimSpace(content),
		},
	}

	u, _ := url.JoinPath(e.url, "/v1/embeddings")

	req, _ := http.NewRequestWithContext(ctx, "POST", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	resp, err := e.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result EmbeddingList

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, errors.New("no embeddings found")
	}

	return &provider.Embedding{
		Data: result.Data[0].Embedding,
	}, nil
}

type EmbeddingList struct {
	Object string `json:"object"`

	Model string      `json:"model"`
	Data  []Embedding `json:"data"`
}

type Embedding struct {
	Object string `json:"object"`

	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}
