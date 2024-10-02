package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"golang.org/x/time/rate"
)

type Reranker interface {
	Limiter
	provider.Reranker
}

type limitedReranker struct {
	limiter  *rate.Limiter
	provider provider.Reranker
}

func NewReranker(l *rate.Limiter, p provider.Reranker) Reranker {
	return &limitedReranker{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedReranker) limiterSetup() {
}

func (p *limitedReranker) Rerank(ctx context.Context, query string, inputs []string, options *provider.RerankOptions) ([]provider.Ranking, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Rerank(ctx, query, inputs, options)
}
