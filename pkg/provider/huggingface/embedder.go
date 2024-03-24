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

	return nil, errors.New("unable to embed input")
}

// func (e *Embedder) Embed(ctx context.Context, content string) ([]float32, error) {
// 	body := EmbeddingsRequest{
// 		Input: strings.TrimSpace(content),
// 	}

// 	url, _ := url.JoinPath(e.url, "/embeddings")
// 	resp, err := e.client.Post(url, "application/json", jsonReader(body))

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New("unable to encode input")
// 	}

// 	var result EmbeddingList

// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	if len(result.Data) == 0 {
// 		return nil, errors.New("unable to embed input")
// 	}

// 	return result.Data[0].Embedding, nil
// }

// type EmbeddingsRequest struct {
// 	Input any    `json:"input"`
// 	Model string `json:"model"`
// }

// type Embedding struct {
// 	Object string `json:"object"` // "embedding"

// 	Index     int       `json:"index"`
// 	Embedding []float32 `json:"embedding"`
// }

// type EmbeddingList struct {
// 	Object string `json:"object"` // "list"

// 	Model string      `json:"model"`
// 	Data  []Embedding `json:"data"`

// 	// usage {
// 	//   prompt_tokens int
// 	//   total_tokens int
// 	// }
// }
