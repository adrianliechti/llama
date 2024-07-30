package index

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider interface {
	List(ctx context.Context, options *ListOptions) ([]Document, error)

	Index(ctx context.Context, documents ...Document) error
	Delete(ctx context.Context, ids ...string) error

	Query(ctx context.Context, query string, options *QueryOptions) ([]Result, error)
}

type ListOptions struct {
}

type QueryOptions struct {
	Limit *int

	Filters map[string]string
}

type Document struct {
	ID string

	Title    string
	Content  string
	Location string

	Metadata map[string]string

	Embedding []float32
}

type Result struct {
	Document
	Score float32
}

type Embedder = provider.Embedder
