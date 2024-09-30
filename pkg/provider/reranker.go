package provider

import (
	"context"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, inputs []string, options *RerankOptions) ([]Ranking, error)
}

type RerankOptions struct {
	Limit *int
}

type Ranking struct {
	Content string
	Score   float64
}
