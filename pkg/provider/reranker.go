package provider

import (
	"context"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, inputs []string, options *RerankOptions) ([]Result, error)
}

type RerankOptions struct {
	Limit *int
}
