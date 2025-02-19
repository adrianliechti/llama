package provider

import (
	"context"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, texts []string, options *RerankOptions) ([]Ranking, error)
}

type RerankOptions struct {
	Limit *int
}

type Ranking struct {
	Text  string
	Score float64
}
