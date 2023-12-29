package rag

import (
	"context"
	"errors"

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

func (*Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	return nil, errors.ErrUnsupported
}
