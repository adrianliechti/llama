package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/extractor"
)

var _ extractor.Provider = &Client{}

type Client struct {
	client *http.Client

	url   string
	token string
}

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		url = "https://r.jina.ai/"
	}

	c := &Client{
		client: http.DefaultClient,

		url: url,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = new(extractor.ExtractOptions)
	}

	if !isSupported(input) {
		return nil, extractor.ErrUnsupported
	}

	// body := map[string]any{
	// 	"url": input.URL,
	// }

	// req, _ := http.NewRequestWithContext(ctx, "POST", c.url, jsonReader(body))
	// req.Header.Set("Content-Type", "application/json")

	url, _ := url.JoinPath(c.url, "/"+input.URL)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("X-Return-Format", "markdown")

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

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return &extractor.Document{
		Content: string(data),
	}, nil
}

func isSupported(input extractor.File) bool {
	if input.URL == "" {
		return false
	}

	return true
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
