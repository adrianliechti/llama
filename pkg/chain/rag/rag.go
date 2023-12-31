package rag

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Completer = &Provider{}
)

type Provider struct {
	index index.Provider

	embedder  provider.Embedder
	completer provider.Completer

	topK int
	topP float32
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

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

func WithTopK(val int) Option {
	return func(p *Provider) {
		p.topK = val
	}
}

func WithTopP(val float32) Option {
	return func(p *Provider) {
		p.topP = val
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	embedding, err := p.index.Embed(ctx, message.Content)

	if err != nil {
		return nil, err
	}

	results, err := p.index.Search(ctx, embedding, &index.SearchOptions{
		TopP: p.topP,
		TopK: p.topK,
	})

	if err != nil {
		return nil, err
	}

	var prompt strings.Builder

	prompt.WriteString(message.Content)

	if len(results) > 0 {
		prompt.WriteString("\n\n")
		prompt.WriteString("Here is some possibly useful information:")

		for _, result := range results {
			prompt.WriteString("\n\n")
			prompt.WriteString(result.Content)

		}
	}

	prompt.WriteString(message.Content)

	messages[len(messages)-1] = provider.Message{
		Role:    provider.MessageRoleUser,
		Content: prompt.String(),
	}

	return p.completer.Complete(ctx, messages, options)
}
