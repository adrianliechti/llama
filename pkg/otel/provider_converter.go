package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/converter"

	"go.opentelemetry.io/otel"
)

type ObservableConverter interface {
	Observable
	converter.Provider
}

type observableConverter struct {
	name    string
	library string

	provider string

	converter converter.Provider
}

func NewConverter(provider string, p converter.Provider) ObservableConverter {
	library := strings.ToLower(provider)

	return &observableConverter{
		converter: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-partitioner") + "-partitioner",
		library: library,

		provider: provider,
	}
}

func (p *observableConverter) otelSetup() {
}

func (p *observableConverter) Convert(ctx context.Context, input converter.File, options *converter.ConvertOptions) (*converter.Document, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.converter.Convert(ctx, input, options)

	return result, err
}
