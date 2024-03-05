package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	u, _ := url.Parse("https://api.tavily.com/search")

	body := map[string]any{
		"api_key":      c.token,
		"query":        query,
		"search_depth": "basic",
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var data SearchResult

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range data.Results {
		result := index.Result{
			Document: index.Document{
				Title:    r.Title,
				Content:  r.Content,
				Location: r.URL,
			},
		}

		results = append(results, result)
	}

	return results, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func convertError(resp *http.Response) error {
	return errors.New(http.StatusText(resp.StatusCode))
}
