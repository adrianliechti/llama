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

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	c := &Config{
		client: http.DefaultClient,

		url:   url,
		model: model,
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
		Model: e.model,

		Input: []string{
			strings.TrimSpace(content),
		},
	}

	u, _ := url.JoinPath(e.url, "/api/embed")
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
		Data: result.Embeddings[0],

		Usage: &provider.Usage{
			InputTokens: result.PromptEvalCount,
		},
	}, nil
}

type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`

	PromptEvalCount int `json:"prompt_eval_count,omitempty"`
}
