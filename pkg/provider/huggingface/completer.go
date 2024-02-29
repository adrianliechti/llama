package huggingface

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config

	client *openai.Completer
}

func NewCompleter(url string, options ...Option) (*Completer, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	url = strings.TrimRight(url, "/")

	if !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}

	cfg := &Config{
		url: url,
	}

	for _, option := range options {
		option(cfg)
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

	return &Completer{
		Config: cfg,
		client: client,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	return c.client.Complete(ctx, messages, options)
}
