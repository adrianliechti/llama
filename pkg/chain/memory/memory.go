package memory

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool/memory"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	completer provider.Completer
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

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if len(memory.Claims) > 0 {

		message := messages[len(messages)-1]

		if message.Role == provider.MessageRoleUser {
			content := message.Content

			content += "\n\nMemorized Claims. Use this information if helpful:\n"

			for _, claim := range memory.Claims {
				content += "  - " + claim + "\n"
			}

			message.Content = content
		}

		messages[len(messages)-1] = message
	}

	return c.completer.Complete(ctx, messages, options)
}
