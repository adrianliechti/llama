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

	embedder  provider.Embedder
	completer provider.Completer

	system string

	limit    *int
	distance *float32

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

	if p.embedder == nil {
		return nil, errors.New("missing embedder provider")
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

func WithEmbedder(embedder provider.Embedder) Option {
	return func(p *Provider) {
		p.embedder = embedder
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

func WithSystem(val string) Option {
	return func(p *Provider) {
		p.system = val
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

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	if p.system != "" {
		if messages[0].Role == provider.MessageRoleSystem {
			messages = messages[1:]
		}

		message := provider.Message{
			Role:    provider.MessageRoleSystem,
			Content: p.system,
		}

		messages = append([]provider.Message{message}, messages...)
	}

	filters := map[string]string{}

	for k, c := range p.filters {
		v, err := c.Categorize(ctx, message.Content)

		if err != nil || v == "" {
			continue
		}

		filters[k] = v
	}

	embedding, err := p.index.Embed(ctx, message.Content)

	if err != nil {
		return nil, err
	}

	results, err := p.index.Query(ctx, embedding, &index.QueryOptions{
		Limit:    p.limit,
		Distance: p.distance,

		Filters: filters,
	})

	if err != nil {
		return nil, err
	}

	var prompt strings.Builder

	if len(results) > 0 {
		prompt.WriteString("You answer questions based on a provided context.\n")
		prompt.WriteString("You answer questions as thoroughly as possible using only the provided context.\n")
		prompt.WriteString("If the context doesn't provide an answer, you indicate that.\n")
		//prompt.WriteString("You provide citations inline with the answer text.\n")
		prompt.WriteString("\n")

		prompt.WriteString("### Context\n")

		for _, result := range results {
			prompt.WriteString("\n\n")
			prompt.WriteString(result.Content)
		}

		prompt.WriteString("### Input\n")
	}

	prompt.WriteString(strings.TrimSpace(message.Content))

	messages[len(messages)-1] = provider.Message{
		Role:    provider.MessageRoleUser,
		Content: prompt.String(),
	}

	return p.completer.Complete(ctx, messages, options)
}
