package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableEmbedder interface {
	Observable
	provider.Embedder
}

type embedder struct {
	name    string
	library string

	model    string
	provider string

	embedder provider.Embedder

	embeddingMeter metric.Int64Counter
}

func NewEmbedder(provider, model string, p provider.Embedder) ObservableEmbedder {
	library := strings.ToLower(provider)

	embeddingMeter, _ := otel.Meter(library).Int64Counter("llm_platform_embedding")

	return &embedder{
		embedder: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-embedder") + "-embedder",
		library: library,

		model:    model,
		provider: provider,

		embeddingMeter: embeddingMeter,
	}
}

func (p *embedder) otelSetup() {
}

func (p *embedder) Embed(ctx context.Context, content string) (provider.Embeddings, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.embedder.Embed(ctx, content)

	p.embeddingMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	if content != "" {
		span.SetAttributes(attribute.String("text", content))
	}

	return result, err
}
