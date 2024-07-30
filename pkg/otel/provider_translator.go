package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableTranslator interface {
	Observable
	provider.Translator
}

type translator struct {
	name    string
	library string

	model    string
	provider string

	translator provider.Translator

	embeddingMeter metric.Int64Counter
}

func NewTranslator(provider, model string, p provider.Translator) ObservableTranslator {
	library := strings.ToLower(provider)

	embeddingMeter, _ := otel.Meter(library).Int64Counter("llm_platform_translation")

	return &translator{
		translator: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-translator") + "-translator",
		library: library,

		model:    model,
		provider: provider,

		embeddingMeter: embeddingMeter,
	}
}

func (p *translator) otelSetup() {
}

func (p *translator) Translate(ctx context.Context, content string, options *provider.TranslateOptions) (*provider.Translation, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.translator.Translate(ctx, content, options)

	p.embeddingMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	if content != "" {
		span.SetAttributes(attribute.String("text", content))
	}

	return result, err
}
