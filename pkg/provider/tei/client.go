package tei

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Embedder = (*Client)(nil)
)

type Client struct {
	url string

	client *http.Client
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	p := &Client{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" {
		return nil, errors.New("invalid url")
	}

	return p, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) Embed(ctx context.Context, content string) ([]float32, error) {
	body := map[string]any{
		"inputs": strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(c.url, "/embed")
	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to encode input")
	}

	defer resp.Body.Close()

	var result []float32

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}