package index

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider interface {
	Embedder

	Index(ctx context.Context, documents ...Document) error
	Search(ctx context.Context, embedding []float32, options *SearchOptions) ([]Result, error)
}

type SearchOptions struct {
	TopK int
	TopP float32
}

type Document struct {
	ID string

	Embedding []float32

	Content  string
	Metadata map[string]any
}

type Result struct {
	Document
	Distance float32
}

type Embedder = provider.Embedder
