package jina

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/reranker"
)

var _ reranker.Provider = (*Client)(nil)

type Client struct {
	url string

	token string
	model string

	client *http.Client
}

func New(url, model string, options ...Option) (*Client, error) {
	if url == "" {
		url = "https://api.jina.ai"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	c := &Client{
		client: http.DefaultClient,

		url: url,

		model: "jina-reranker-v2-base-multilingual",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Rerank(ctx context.Context, query string, inputs []string, options *reranker.RerankOptions) ([]reranker.Result, error) {
	if options == nil {
		options = new(reranker.RerankOptions)
	}

	body := map[string]any{
		"model": c.model,

		"query":     query,
		"documents": inputs,
	}

	if options.Limit != nil {
		body["top_n"] = *options.Limit
	}

	u, _ := url.JoinPath(c.url, "/v1/rerank")

	req, _ := http.NewRequestWithContext(ctx, "POST", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)

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

	var result []reranker.Result

	for _, r := range data.Results {
		result = append(result, reranker.Result{
			Content: inputs[r.Index],
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
	Score    float64  `json:"relevance_score"`
}

type Document struct {
	Text string `json:"text"`
}
