package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"go.opentelemetry.io/otel"
)

type ObservableExtractor interface {
	Observable
	extractor.Provider
}

type observableExtractor struct {
	name    string
	library string

	provider string

	extractor extractor.Provider
}

func NewExtractor(provider string, p extractor.Provider) ObservableExtractor {
	library := strings.ToLower(provider)

	return &observableExtractor{
		extractor: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-partitioner") + "-partitioner",
		library: library,

		provider: provider,
	}
}

func (p *observableExtractor) otelSetup() {
}

func (p *observableExtractor) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.extractor.Extract(ctx, input, options)

	return result, err
}
