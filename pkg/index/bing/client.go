package bing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/wingman/pkg/index"
)

var _ index.Provider = &Client{}

type Client struct {
	token  string
	client *http.Client
}

func New(token string, options ...Option) (*Client, error) {
	c := &Client{
		token:  token,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	if c.token == "" {
		return nil, errors.New("invalid token")
	}

	return c, nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	u, _ := url.Parse("https://api.bing.microsoft.com/v7.0/search")

	values := u.Query()
	values.Set("q", query)

	u.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data response

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, p := range data.WebPages.Value {
		result := index.Result{
			Document: index.Document{
				Title:   p.Name,
				Source:  p.URL,
				Content: p.Snippet,
			},
		}

		results = append(results, result)
	}

	return results, nil
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Document], error) {
	return nil, errors.ErrUnsupported
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	return errors.ErrUnsupported
}

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	return errors.ErrUnsupported
}
