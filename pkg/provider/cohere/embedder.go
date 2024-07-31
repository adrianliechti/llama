package cohere

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
}

func NewEmbedder(options ...Option) (*Embedder, error) {
	cfg := &Config{
		url: "https://api.cohere.com",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	url, _ := url.JoinPath(e.url, "/v1/embed")

	body := map[string]any{
		"model": e.model,

		"texts": []string{
			content,
		},

		"input_type": "search_document",

		"embedding_types": []string{
			"float",
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
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

	var result EmbedResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	floats := result.Embeddings["float"]

	if len(floats) == 0 {
		return nil, errors.New("invalid embeddings")
	}

	return &provider.Embedding{
		Data: floats[0],
	}, nil
}

type EmbedResponse struct {
	Embeddings map[string][][]float32 `json:"embeddings"`
}
