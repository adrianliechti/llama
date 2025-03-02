package limiter

import (
	"context"

	"github.com/adrianliechti/wingman/pkg/provider"

	"golang.org/x/time/rate"
)

type Completer interface {
	Limiter
	provider.Completer
}

type limitedCompleter struct {
	limiter  *rate.Limiter
	provider provider.Completer
}

func NewCompleter(l *rate.Limiter, p provider.Completer) Completer {
	return &limitedCompleter{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedCompleter) limiterSetup() {
}

func (p *limitedCompleter) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Complete(ctx, messages, options)
}
