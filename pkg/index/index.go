package index

import (
	"context"
)

type Index interface {
	Index(ctx context.Context, documents ...Document) error
	Search(ctx context.Context, embedding []float32) ([]Result, error)
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
