package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/wingman/pkg/segmenter"
	"go.opentelemetry.io/otel"
)

type Segmenter interface {
	Observable
	segmenter.Provider
}

type observableSegmenter struct {
	name    string
	library string

	provider string

	segmenter segmenter.Provider
}

func NewSegmenter(provider string, p segmenter.Provider) Segmenter {
	library := strings.ToLower(provider)

	return &observableSegmenter{
		segmenter: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-segmenter") + "-segmenter",
		library: library,

		provider: provider,
	}
}

func (p *observableSegmenter) otelSetup() {
}

func (p *observableSegmenter) Segment(ctx context.Context, input string, options *segmenter.SegmentOptions) ([]segmenter.Segment, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.segmenter.Segment(ctx, input, options)

	return result, err
}
