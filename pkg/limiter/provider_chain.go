package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
	"golang.org/x/time/rate"
)

type Chain interface {
	Limiter
	chain.Provider
}

type limitedChain struct {
	limiter  *rate.Limiter
	provider chain.Provider
}

func NewChain(l *rate.Limiter, p chain.Provider) Chain {
	return &limitedChain{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedChain) limiterSetup() {
}

func (p *limitedChain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Complete(ctx, messages, options)
}
