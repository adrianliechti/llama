package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Embedder interface {
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

func NewEmbedder(provider, model string, p provider.Embedder) Embedder {
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

func (p *observableEmbedder) Embed(ctx context.Context, texts []string) (*provider.Embedding, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.embedder.Embed(ctx, texts)

	meterRequest(ctx, p.library, p.provider, "embed", p.model)

	if EnableDebug {
		span.SetAttributes(attribute.StringSlice("texts", texts))
	}

	if result != nil {
		if result.Usage != nil {
			tokens := int64(result.Usage.InputTokens) + int64(result.Usage.OutputTokens)
			meterTokens(ctx, p.library, p.provider, "embed", p.model, tokens)
		}
	}

	return result, err
}
