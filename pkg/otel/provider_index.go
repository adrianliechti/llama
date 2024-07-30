package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableIndex interface {
	Observable
	index.Provider
}

type indexer struct {
	name    string
	library string

	provider string

	index index.Provider

	queryMeter metric.Int64Counter
}

func NewIndex(provider, index string, p index.Provider) ObservableIndex {
	library := strings.ToLower(provider)

	queryMeter, _ := otel.Meter(library).Int64Counter("llm_platform_query")

	return &indexer{
		index: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-index") + "-index",
		library: library,

		provider: provider,

		queryMeter: queryMeter,
	}
}

func (p *indexer) otelSetup() {
}

func (p *indexer) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.index.List(ctx, options)

	return result, err
}

func (p *indexer) Index(ctx context.Context, documents ...index.Document) error {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	err := p.index.Index(ctx, documents...)

	return err
}

func (p *indexer) Delete(ctx context.Context, ids ...string) error {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	err := p.index.Delete(ctx, ids...)

	return err
}

func (p *indexer) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	p.queryMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
	))

	result, err := p.index.Query(ctx, query, options)

	if result != nil {
		var outputs []string

		for _, r := range result {
			outputs = append(outputs, r.Content)
		}

		if len(outputs) > 0 {
			span.SetAttributes(attribute.StringSlice("output", outputs))
		}
	}

	return result, err
}
