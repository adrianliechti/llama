package duckduckgo

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/text"
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

// Query implements index.Provider.
func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	url, _ := url.Parse("https://duckduckgo.com/html/")

	values := url.Query()
	values.Set("q", query)

	url.RawQuery = values.Encode()

	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("Referer", "https://www.duckduckgo.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Safari/605.1.15")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var results []index.Result

	re := regexp.MustCompile(`<[^>]*>`)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, "result__snippet") {
			continue
		}

		snippet := re.ReplaceAllString(line, "")
		snippet = text.Normalize(snippet)

		result := index.Result{
			Document: index.Document{
				Content: snippet,
			},
		}

		results = append(results, result)
	}

	return results, nil
}
