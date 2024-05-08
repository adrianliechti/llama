package ollama

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
		url = "http://localhost:11434"
	}

	c := &Config{
		url:    url,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	go c.ensureModel()

	return &Embedder{
		Config: c,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (provider.Embeddings, error) {
	body := &EmbeddingRequest{
		Model:  e.model,
		Prompt: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(e.url, "/api/embeddings")
	resp, err := e.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to embed")
	}

	var result EmbeddingResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return toFloat32s(result.Embedding), nil
}

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}
