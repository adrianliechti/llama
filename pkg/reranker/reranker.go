package reranker

import "context"

type Provider interface {
	Rerank(ctx context.Context, query string, inputs []string, options *RerankOptions) ([]Result, error)
}

type RerankOptions struct {
	Limit *int
}

type Result struct {
	Content string
	Score   float64
}
