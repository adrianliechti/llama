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

	if model == "" {
		model = "jina-reranker-v2-base-multilingual"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{
		url:   url,
		model: model,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Reranker{
		Config: cfg,
	}, nil
}

func (r *Reranker) Rerank(ctx context.Context, query string, texts []string, options *provider.RerankOptions) ([]provider.Ranking, error) {
	if options == nil {
		options = new(provider.RerankOptions)
	}

	body := map[string]any{
		"model": r.model,

		"query":     query,
		"documents": texts,
	}

	if options.Limit != nil {
		body["top_n"] = *options.Limit
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

	var result []provider.Ranking

	for _, r := range data.Results {
		result = append(result, provider.Ranking{
			Text:  texts[r.Index],
			Score: r.Score,
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
	Score    float64  `json:"relevance_score"`
}

type Document struct {
	Text string `json:"text"`
}
