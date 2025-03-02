package searxng

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
	url    string
	client *http.Client
}

func New(url string, options ...Option) (*Client, error) {
	c := &Client{
		url:    url,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	url, _ := url.Parse(c.url)
	url = url.JoinPath("/search")

	values := url.Query()
	values.Set("q", query)
	values.Set("format", "json")
	values.Set("safesearch", "0")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch search results")
	}

	var data searchResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range data.Results {
		result := index.Result{
			Document: index.Document{
				Title:   r.Title,
				Source:  r.URL,
				Content: r.Content,
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
