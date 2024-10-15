package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/extractor"

	"golang.org/x/time/rate"
)

type Extractor interface {
	Limiter
	extractor.Provider
}

type limitedExtractor struct {
	limiter  *rate.Limiter
	provider extractor.Provider
}

func NewExtractor(l *rate.Limiter, p extractor.Provider) Extractor {
	return &limitedExtractor{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedExtractor) limiterSetup() {
}

func (p *limitedExtractor) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Extract(ctx, input, options)
}
