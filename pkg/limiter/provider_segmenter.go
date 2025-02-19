package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/segmenter"

	"golang.org/x/time/rate"
)

type Segmenter interface {
	Limiter
	segmenter.Provider
}

type limitedSegmenter struct {
	limiter  *rate.Limiter
	provider segmenter.Provider
}

func NewSegmenter(l *rate.Limiter, p segmenter.Provider) Segmenter {
	return &limitedSegmenter{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedSegmenter) limiterSetup() {
}

func (p *limitedSegmenter) Segment(ctx context.Context, input string, options *segmenter.SegmentOptions) ([]segmenter.Segment, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Segment(ctx, input, options)
}
