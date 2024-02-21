package tgi

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

var (
	_ provider.Completer = (*Client)(nil)
)

type Client struct {
	url string

	client *openai.Completer
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	url = strings.TrimRight(url, "/")

	if !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}

	p := &Client{
		url: url,
	}

	for _, option := range options {
		option(p)
	}

	opts := []openai.Option{
		openai.WithURL(url),
		openai.WithToken("-"),
		openai.WithModel("tgi"),
	}

	client, err := openai.NewCompleter(opts...)

	if err != nil {
		return nil, err
	}

	p.client = client

	return p, nil
}

func (c *Client) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	return c.client.Complete(ctx, messages, options)
}
