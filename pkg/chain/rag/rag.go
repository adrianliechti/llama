package rag

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = &Provider{}

type Provider struct {
	index index.Provider

	prompt *prompt.Prompt

	limit    *int
	distance *float32

	completer      provider.Completer
	contextualizer provider.Completer

	filters map[string]classifier.Provider
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{
		prompt: prompt.MustNew(promptTemplate),

		filters: map[string]classifier.Provider{},
	}

	for _, option := range options {
		option(p)
	}

	if p.index == nil {
		return nil, errors.New("missing index provider")
	}

	if p.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return p, nil
}

func WithIndex(index index.Provider) Option {
	return func(p *Provider) {
		p.index = index
	}
}

func WithPrompt(prompt *prompt.Prompt) Option {
	return func(p *Provider) {
		p.prompt = prompt
	}
}

func WithLimit(val int) Option {
	return func(p *Provider) {
		p.limit = &val
	}
}

func WithDistance(val float32) Option {
	return func(p *Provider) {
		p.distance = &val
	}
}

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func WithContextualizer(val provider.Completer) Option {
	return func(p *Provider) {
		p.contextualizer = val
	}
}

func WithFilter(name string, classifier classifier.Provider) Option {
	return func(p *Provider) {
		p.filters[name] = classifier
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	if p.contextualizer != nil && len(messages) > 1 {
		result, err := p.contextualizer.Complete(ctx, messages, nil)

		if err != nil {
			return nil, err
		}

		message = provider.Message{
			Role:    provider.MessageRoleUser,
			Content: strings.TrimSpace(result.Message.Content),
		}

		messages = []provider.Message{message}
	}

	filters := map[string]string{}

	for k, c := range p.filters {
		v, err := c.Categorize(ctx, message.Content)

		if err != nil || v == "" {
			continue
		}

		filters[k] = v
	}

	results, err := p.index.Query(ctx, message.Content, &index.QueryOptions{
		Limit:    p.limit,
		Distance: p.distance,

		Filters: filters,
	})

	if err != nil {
		return nil, err
	}

	data := promptData{
		Input: strings.TrimSpace(message.Content),
	}

	for _, r := range results {
		data.Results = append(data.Results, promptResult{
			Metadata: r.Metadata,
			Content:  strings.TrimSpace(r.Content),
		})
	}

	prompt, err := p.prompt.Execute(data)

	if err != nil {
		return nil, err
	}

	println(prompt)

	message = provider.Message{
		Role:    provider.MessageRoleUser,
		Content: prompt,
	}

	messages[len(messages)-1] = message

	return p.completer.Complete(ctx, messages, options)
}
