package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ObservableIndex interface {
	Observable
	index.Provider
}

type observableIndex struct {
	name    string
	library string

	provider string

	index index.Provider
}

func NewIndex(provider, index string, p index.Provider) ObservableIndex {
	library := strings.ToLower(provider)

	return &observableIndex{
		index: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-index") + "-index",
		library: library,

		provider: provider,
	}
}

func (p *observableIndex) otelSetup() {
}

func (p *observableIndex) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.index.List(ctx, options)

	return result, err
}

func (p *observableIndex) Index(ctx context.Context, documents ...index.Document) error {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	err := p.index.Index(ctx, documents...)

	return err
}

func (p *observableIndex) Delete(ctx context.Context, ids ...string) error {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	err := p.index.Delete(ctx, ids...)

	return err
}

func (p *observableIndex) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.index.Query(ctx, query, options)

	if EnableDebug {
		span.SetAttributes(attribute.String("query", query))

		if result != nil {
			var outputs []string

			for _, r := range result {
				outputs = append(outputs, r.Content)
			}

			if len(outputs) > 0 {
				span.SetAttributes(attribute.StringSlice("results", outputs))
			}
		}
	}

	return result, err
}
