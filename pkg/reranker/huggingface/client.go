package huggingface

import (
	"context"
	"errors"
	"net/http"
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
		url = "https://api-inference.huggingface.co/models/" + model
	}

	url = strings.TrimRight(url, "/")

	c := &Client{
		client: http.DefaultClient,

		url:   url,
		token: "-",

		model: "tei",
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
		"query": strings.TrimSpace(query),
		"texts": inputs,
	}

	if strings.Contains(c.url, "api-inference.huggingface.co") {
		body = map[string]any{
			"source_sentence": query,
			"sentences":       inputs,
		}
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.url, jsonReader(body))
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

	return nil, errors.New("unable to rerank input")
}
