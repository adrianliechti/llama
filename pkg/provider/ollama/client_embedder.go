package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	url   string
	model string

	client *http.Client
}

func NewEmbedder(url string, options ...Option) (*Embedder, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	c := &Config{
		options: []openai.Option{
			openai.WithURL(url + "/v1"),
		},
	}

	for _, option := range options {
		option(c)
	}

	cfg := &openai.Config{}

	for _, option := range c.options {
		option(cfg)
	}

	e := &Embedder{
		url:   strings.TrimSuffix(url, "/v1"),
		model: cfg.Model,

		client: http.DefaultClient,
	}

	return e, nil
}

// func NewEmbedder(url string, options ...Option) (*openai.Embedder, error) {
// 	if url == "" {
// 		url = "http://localhost:11434"
// 	}

// 	url = strings.TrimRight(url, "/")
// 	url = strings.TrimSuffix(url, "/v1")

// 	c := &Config{
// 		options: []openai.Option{
// 			openai.WithURL(url + "/v1"),
// 		},
// 	}

// 	for _, option := range options {
// 		option(c)
// 	}

// 	return openai.NewEmbedder(c.options...)
// }

func (e *Embedder) Embed(ctx context.Context, content string) ([]float32, error) {
	body := &EmbeddingRequest{
		Model:  e.model,
		Prompt: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(e.url, "/api/embeddings")
	resp, err := e.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to embed")
	}

	defer resp.Body.Close()

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
