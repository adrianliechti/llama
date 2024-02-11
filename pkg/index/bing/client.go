package bing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
)

var (
	_ index.Provider = (*Client)(nil)
)

type Client struct {
	client *http.Client

	token string
}

type Option func(*Client)

func New(token string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		token: token,
	}

	for _, option := range options {
		option(c)
	}

	if c.token == "" {
		return nil, errors.New("invalid token")
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	return errors.ErrUnsupported
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	return nil, errors.ErrUnsupported
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	url, _ := url.Parse("https://api.bing.microsoft.com/v7.0/search")

	values := url.Query()
	values.Set("q", query)

	url.RawQuery = values.Encode()

	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, page := range data.WebPages.Value {
		result := index.Result{
			Document: index.Document{
				ID: page.ID,

				URL:     page.URL,
				Content: page.Snippet,
			},
		}

		results = append(results, result)
	}

	return results, nil
}
