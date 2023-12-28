package rag

import (
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider struct {
	index index.Index

	embedder  provider.Embedder
	completer provider.Completer
}

func New(index index.Index, embedder provider.Embedder, completer provider.Completer) *Provider {
	p := &Provider{
		index: index,

		embedder:  embedder,
		completer: completer,
	}

	return p
}
