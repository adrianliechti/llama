package ollama

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	completer *openai.Completer
}

func NewCompleter(url, model string, options ...Option) (*Completer, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	opts := []openai.Option{}

	if cfg.client != nil {
		opts = append(opts, openai.WithClient(cfg.client))
	}

	completer, err := openai.NewCompleter(url+"/v1", model, opts...)

	if err != nil {
		return nil, err
	}

	return &Completer{
		completer: completer,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	inputOptions := &provider.CompleteOptions{
		Stream: options.Stream,

		Stop:  options.Stop,
		Tools: options.Tools,

		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,

		Format: options.Format,
	}

	if len(options.Tools) > 0 {
		inputOptions.Stream = nil
	}

	result, err := c.completer.Complete(ctx, messages, inputOptions)

	if err != nil {
		return nil, err
	}

	if inputOptions.Stream == nil && options.Stream != nil {
		if err := options.Stream(ctx, *result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
