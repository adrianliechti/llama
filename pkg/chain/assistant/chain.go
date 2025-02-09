package assistant

import (
	"context"
	"errors"
	"slices"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/template"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	completer provider.Completer

	messages []provider.Message

	effort      provider.ReasoningEffort
	temperature *float32
}

type Option func(*Chain)

func New(options ...Option) (*Chain, error) {
	c := &Chain{}

	for _, option := range options {
		option(c)
	}

	if c.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return c, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(c *Chain) {
		c.completer = completer
	}
}

func WithMessages(messages ...provider.Message) Option {
	return func(c *Chain) {
		c.messages = messages
	}
}

func WithEffort(effort provider.ReasoningEffort) Option {
	return func(c *Chain) {
		c.effort = effort
	}
}

func WithTemperature(temperature float32) Option {
	return func(c *Chain) {
		c.temperature = &temperature
	}
}

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if options.Effort == "" {
		options.Effort = c.effort
	}

	if options.Temperature == nil {
		options.Temperature = c.temperature
	}

	if len(c.messages) > 0 {
		values, err := template.Messages(c.messages, nil)

		if err != nil {
			return nil, err
		}

		messages = slices.Concat(values, messages)
	}

	return c.completer.Complete(ctx, messages, options)
}
