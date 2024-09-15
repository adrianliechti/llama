package huggingface

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Reranker = (*Reranker)(nil)

type Reranker struct {
	*Config
}

func NewReranker(url string, options ...Option) (*Reranker, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	cfg := &Config{
		client: http.DefaultClient,

		url:   strings.TrimRight(url, "/"),
		token: "-",

		model: "tei",
	}

	for _, option := range options {
		option(cfg)
	}

	return &Reranker{
		Config: cfg,
	}, nil
}

func (r *Reranker) Rerank(ctx context.Context, query string, inputs []string) ([]provider.Result, error) {
	body := map[string]any{
		"query": strings.TrimSpace(query),
		"texts": inputs,
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", r.url, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	return nil, errors.New("unable to embed input")
}
