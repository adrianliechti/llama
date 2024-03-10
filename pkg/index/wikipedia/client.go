package wikipedia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/adrianliechti/llama/pkg/index"
)

var (
	_ index.Provider = (*Client)(nil)
)

type Client struct {
	client *http.Client
}

type Option func(*Client)

func New(options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,
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

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	return errors.ErrUnsupported
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	return nil, errors.ErrUnsupported
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	pages, err := c.search(ctx, query)

	if err != nil {
		return nil, err
	}

	var results []index.Result

	for _, p := range pages {
		page, err := c.page(ctx, p.ID)

		if err != nil {
			return nil, err
		}

		results = append(results, index.Result{
			Document: index.Document{
				ID: fmt.Sprintf("%d", page.ID),

				Title:   page.Title,
				Content: page.Extract,
			},
		})
	}

	return results, nil
}

func (c *Client) search(ctx context.Context, query string) ([]Search, error) {
	url, _ := url.Parse("https://en.wikipedia.org/w/api.php")

	values := url.Query()
	values.Set("action", "query")
	values.Set("list", "search")
	values.Set("srsearch", query)
	values.Set("utf8", "")
	values.Set("format", "json")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to search wikipedia")
	}

	var data SearchQuery

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Query.Search, nil
}

func (c *Client) page(ctx context.Context, pageID int) (*Page, error) {
	url, _ := url.Parse("https://en.wikipedia.org/w/api.php")

	values := url.Query()
	values.Set("action", "query")
	values.Set("prop", "extracts")
	values.Set("exintro", "")
	values.Set("explaintext", "")
	values.Set("pageids", fmt.Sprintf("%d", pageID))
	values.Set("utf8", "")
	values.Set("format", "json")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to get wikipedia page")
	}

	var data PageQuery

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, page := range data.Query.Pages {
		return &page, nil
	}

	return nil, errors.New("page not found")
}

type SearchQuery struct {
	Query struct {
		Search []Search `json:"search,omitempty"`
	} `json:"query,omitempty"`
}

type Search struct {
	NS        int       `json:"ns,omitempty"`
	ID        int       `json:"pageid,omitempty"`
	Title     string    `json:"title,omitempty"`
	Size      int       `json:"size,omitempty"`
	Words     int       `json:"wordcount,omitempty"`
	Snippet   string    `json:"snippet,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type PageQuery struct {
	Query struct {
		Pages map[string]Page `json:"pages,omitempty"`
	} `json:"query,omitempty"`
}

type Page struct {
	NS      int    `json:"ns,omitempty"`
	ID      int    `json:"pageid,omitempty"`
	Title   string `json:"title,omitempty"`
	Extract string `json:"extract,omitempty"`
}
