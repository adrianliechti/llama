package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableRenderer interface {
	Observable
	provider.Renderer
}

type renderer struct {
	name    string
	library string

	model    string
	provider string

	renderer provider.Renderer

	embeddingMeter metric.Int64Counter
}

func NewRenderer(provider, model string, p provider.Renderer) ObservableRenderer {
	library := strings.ToLower(provider)

	embeddingMeter, _ := otel.Meter(library).Int64Counter("llm_platform_image")

	return &renderer{
		renderer: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-renderer") + "-renderer",
		library: library,

		model:    model,
		provider: provider,

		embeddingMeter: embeddingMeter,
	}
}

func (p *renderer) otelSetup() {
}

func (p *renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.renderer.Render(ctx, input, options)

	p.embeddingMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	return result, err
}
