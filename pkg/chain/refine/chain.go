package refine

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	index index.Provider

	prompt *prompt.Prompt

	limit    *int
	distance *float32

	completer      provider.Completer
	contextualizer provider.Completer

	filters map[string]classifier.Provider
}

type Option func(*Chain)

func New(options ...Option) (*Chain, error) {
	c := &Chain{
		prompt: prompt.MustNew(promptTemplate),

		filters: map[string]classifier.Provider{},
	}

	for _, option := range options {
		option(c)
	}

	if c.index == nil {
		return nil, errors.New("missing index provider")
	}

	if c.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return c, nil
}

func WithIndex(index index.Provider) Option {
	return func(c *Chain) {
		c.index = index
	}
}

func WithPrompt(prompt *prompt.Prompt) Option {
	return func(c *Chain) {
		c.prompt = prompt
	}
}

func WithLimit(val int) Option {
	return func(c *Chain) {
		c.limit = &val
	}
}

func WithDistance(val float32) Option {
	return func(c *Chain) {
		c.distance = &val
	}
}

func WithCompleter(completer provider.Completer) Option {
	return func(c *Chain) {
		c.completer = completer
	}
}

func WithContextualizer(val provider.Completer) Option {
	return func(c *Chain) {
		c.contextualizer = val
	}
}

func WithFilter(name string, classifier classifier.Provider) Option {
	return func(c *Chain) {
		c.filters[name] = classifier
	}
}

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	if c.contextualizer != nil && len(messages) > 1 {
		result, err := c.contextualizer.Complete(ctx, messages, nil)

		if err != nil {
			return nil, err
		}

		message = provider.Message{
			Role:    provider.MessageRoleUser,
			Content: strings.TrimSpace(result.Message.Content),
		}

		//messages = []provider.Message{message}
	}

	filters := map[string]string{}

	for k, c := range c.filters {
		v, err := c.Classify(ctx, message.Content)

		if err != nil || v == "" {
			continue
		}

		filters[k] = v
	}

	results, err := c.index.Query(ctx, message.Content, &index.QueryOptions{
		Limit:    c.limit,
		Distance: c.distance,

		Filters: filters,
	})

	if err != nil {
		return nil, err
	}

	var answer string
	var result *provider.Completion

	for _, r := range results {
		data := promptData{
			Input:  strings.TrimSpace(message.Content),
			Answer: answer,

			Results: []promptResult{
				{
					Metadata: r.Metadata,
					Content:  strings.TrimSpace(r.Content),
				},
			},
		}

		prompt, err := c.prompt.Execute(data)

		if err != nil {
			return nil, err
		}

		println(prompt)

		m := provider.Message{
			Role:    provider.MessageRoleUser,
			Content: prompt,
		}

		completion, err := c.completer.Complete(ctx, []provider.Message{m}, nil)

		if err != nil {
			return nil, err
		}

		answer = strings.TrimSpace(completion.Message.Content)
		result = completion
	}

	return result, nil
}
