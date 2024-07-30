package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ObservableEmbedder interface {
	Observable
	provider.Embedder
}

type observableEmbedder struct {
	name    string
	library string

	model    string
	provider string

	embedder provider.Embedder
}

func NewEmbedder(provider, model string, p provider.Embedder) ObservableEmbedder {
	library := strings.ToLower(provider)

	return &observableEmbedder{
		embedder: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-embedder") + "-embedder",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *observableEmbedder) otelSetup() {
}

func (p *observableEmbedder) Embed(ctx context.Context, content string) (provider.Embeddings, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.embedder.Embed(ctx, content)

	meterRequest(ctx, p.library, p.provider, "embed", p.model)

	if content != "" {
		span.SetAttributes(attribute.String("input", content))
	}

	return result, err
}
