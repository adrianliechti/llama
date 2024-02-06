package index

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider interface {
	List(ctx context.Context, options *ListOptions) ([]Document, error)

	Index(ctx context.Context, documents ...Document) error
	Query(ctx context.Context, query string, options *QueryOptions) ([]Result, error)
}

type ListOptions struct {
}

type QueryOptions struct {
	Limit    *int
	Distance *float32

	Filters map[string]string
}

type Document struct {
	ID string

	Embedding []float32

	Content  string
	Metadata map[string]string
}

type Result struct {
	Document
	Distance float32
}

type Embedder = provider.Embedder
