package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"golang.org/x/time/rate"
)

type Embedder interface {
	Limiter
	provider.Embedder
}

type limitedEmbedder struct {
	limiter  *rate.Limiter
	provider provider.Embedder
}

func NewEmbedder(l *rate.Limiter, p provider.Embedder) Embedder {
	return &limitedEmbedder{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedEmbedder) limiterSetup() {
}

func (p *limitedEmbedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Embed(ctx, content)
}
