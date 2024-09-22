package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/reranker"
	"golang.org/x/time/rate"
)

type Reranker interface {
	Limiter
	reranker.Provider
}

type limitedReranker struct {
	limiter  *rate.Limiter
	provider reranker.Provider
}

func NewReranker(l *rate.Limiter, p reranker.Provider) Reranker {
	return &limitedReranker{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedReranker) limiterSetup() {
}

func (p *limitedReranker) Rerank(ctx context.Context, query string, inputs []string, options *reranker.RerankOptions) ([]reranker.Result, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Rerank(ctx, query, inputs, options)
}
