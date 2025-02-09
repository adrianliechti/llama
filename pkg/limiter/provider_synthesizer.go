package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"golang.org/x/time/rate"
)

type Synthesizer interface {
	Limiter
	provider.Synthesizer
}

type limitedSynthesizer struct {
	limiter  *rate.Limiter
	provider provider.Synthesizer
}

func NewSynthesizer(l *rate.Limiter, p provider.Synthesizer) Synthesizer {
	return &limitedSynthesizer{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedSynthesizer) limiterSetup() {
}

func (p *limitedSynthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Synthesize(ctx, content, options)
}
