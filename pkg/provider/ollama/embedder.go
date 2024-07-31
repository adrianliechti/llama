package ollama

import (
	"context"
	"encoding/json"
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
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return &Embedder{
		Config: c,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
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
		return nil, convertError(resp)
	}

	var result EmbeddingResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &provider.Embedding{
		Data: toFloat32s(result.Embedding),
	}, nil
}

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func toFloat32s(v []float64) []float32 {
	result := make([]float32, len(v))

	for i, x := range v {
		result[i] = float32(x)
	}

	return result
}
