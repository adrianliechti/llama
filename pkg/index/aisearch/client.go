package aisearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"
)

var (
	_ index.Provider = (*Client)(nil)
)

type Client struct {
	client *http.Client

	url   string
	token string

	namespace string
}

type Option func(*Client)

func New(url, namespace, token string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url:   url,
		token: token,

		namespace: namespace,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	return nil, errors.ErrUnsupported
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	return errors.ErrUnsupported
}

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	return errors.ErrUnsupported
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	if options.Limit == nil {
		options.Limit = to.Ptr(10)
	}

	u, _ := url.JoinPath(c.url, "/indexes/"+c.namespace+"/docs")
	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("search", query)
	values.Set("$top", fmt.Sprintf("%d", *options.Limit))
	//values.Set("queryType", "semantic")
	//values.Set("semanticConfiguration", "my-semantic-config")
	values.Set("api-version", "2023-11-01")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	req.Header.Set("api-key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result Results

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range result.Value {
		result := index.Result{
			Document: index.Document{
				ID: r.ID(),

				Title:    r.Title(),
				Content:  r.Content(),
				Location: r.Location(),
			},
		}

		results = append(results, result)
	}

	return results, nil
}
