package provider

import (
	"context"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, inputs []string) ([]Result, error)
}
