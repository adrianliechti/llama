package rag

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = &Provider{}

type Provider struct {
	index index.Provider

	completer provider.Completer

	limit    *int
	distance *float32

	contextualization provider.Completer

	filters map[string]classifier.Provider
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{
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

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func WithFilter(name string, classifier classifier.Provider) Option {
	return func(p *Provider) {
		p.filters[name] = classifier
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

func WithContextualization(val provider.Completer) Option {
	return func(p *Provider) {
		p.contextualization = val
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	if p.contextualization != nil && len(messages) > 1 {
		result, err := p.contextualization.Complete(ctx, messages, nil)

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
		Input:   strings.TrimSpace(message.Content),
		Results: results,
	}

	prompt, err := promptTemplate.Execute(data)

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
