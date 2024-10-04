package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
}

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	cfg := &Config{
		client: http.DefaultClient,

		url:   "https://generativelanguage.googleapis.com",
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	body := EmbedRequest{
		Model: e.model,

		Content: Content{
			Parts: []ContentPart{
				{
					Text: content,
				},
			},
		},
	}

	url, _ := url.JoinPath(e.url, "/v1beta/models/"+e.model+":embedContent")

	if e.token != "" {
		url += "?key=" + e.token
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
	req.Header.Set("content-type", "application/json")

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

	return &provider.Embedding{
		Data: result.Embedding.Values,
	}, nil
}

type EmbedRequest struct {
	Model   string  `json:"model"`
	Content Content `json:"content"`
}

type EmbedResponse struct {
	Embedding Embedding `json:"embedding"`
}

type Embedding struct {
	Values []float32 `json:"values"`
}
