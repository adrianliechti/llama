package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
}

func NewRenderer(provider, model string, p provider.Renderer) ObservableRenderer {
	library := strings.ToLower(provider)

	return &renderer{
		renderer: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-renderer") + "-renderer",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *renderer) otelSetup() {
}

func (p *renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.renderer.Render(ctx, input, options)

	meterRequest(ctx, p.library, p.provider, p.model, "rendering")

	if input != "" {
		span.SetAttributes(attribute.String("input", input))
	}

	return result, err
}
