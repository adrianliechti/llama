package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Reranker interface {
	Observable
	provider.Reranker
}

type observableReranker struct {
	name    string
	library string

	model    string
	provider string

	reranker provider.Reranker
}

func NewReranker(provider, model string, p provider.Reranker) Reranker {
	library := strings.ToLower(provider)

	return &observableReranker{
		reranker: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-reranker") + "-reranker",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *observableReranker) otelSetup() {
}

func (p *observableReranker) Rerank(ctx context.Context, query string, inputs []string, options *provider.RerankOptions) ([]provider.Ranking, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.reranker.Rerank(ctx, query, inputs, options)

	meterRequest(ctx, p.library, p.provider, "rerank", p.model)

	if EnableDebug {
		span.SetAttributes(attribute.String("query", query))
		span.SetAttributes(attribute.StringSlice("inputs", inputs))
	}

	return result, err
}
