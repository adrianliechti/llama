package rag

import (
	"github.com/adrianliechti/llama/pkg/index"
)

type Provider struct {
	index index.Index
}

func New(index index.Index) *Provider {
	p := &Provider{
		index: index,
	}

	return p
}
