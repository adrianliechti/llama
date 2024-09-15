package reranker

import (
	"context"
)

type Provider interface {
	Rerank(ctx context.Context, query string, inputs []string) ([]Result, error)
}

type Result struct {
	Content string
	Score   float32
}
