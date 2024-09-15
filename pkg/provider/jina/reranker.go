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

var _ provider.Reranker = (*Reranker)(nil)

type Reranker struct {
	*Config
}

func NewReranker(url, model string, options ...Option) (*Reranker, error) {
	if url == "" {
		url = "https://api.jina.ai"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{
		client: http.DefaultClient,

		url: url,

		model: "jina-reranker-v2-base-multilingual",
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
		"query": query,
		"input": inputs,
	}

	u, _ := url.JoinPath(r.url, "/v1/rerank")

	req, _ := http.NewRequestWithContext(ctx, "POST", u, jsonReader(body))
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

	var data RerankList

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Results) == 0 {
		return nil, errors.New("no reranks found")
	}

	var result []provider.Result

	for _, r := range data.Results {
		result = append(result, provider.Result{
			Content: r.Document.Text,
			Score:   r.Score,
		})
	}

	return result, nil
}

type RerankList struct {
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

type Result struct {
	Index int `json:"index"`

	Document Document `json:"document"`
	Score    float32  `json:"relevance_score"`
}

type Document struct {
	Text string `json:"text"`
}
