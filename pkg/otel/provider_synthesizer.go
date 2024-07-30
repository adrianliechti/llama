package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableSynthesizer interface {
	Observable
	provider.Synthesizer
}

type synthesizer struct {
	name    string
	library string

	model    string
	provider string

	synthesizer provider.Synthesizer

	embeddingMeter metric.Int64Counter
}

func NewSynthesizer(provider, model string, p provider.Synthesizer) ObservableSynthesizer {
	library := strings.ToLower(provider)

	embeddingMeter, _ := otel.Meter(library).Int64Counter("llm_platform_audio")

	return &synthesizer{
		synthesizer: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-synthesizer") + "-synthesizer",
		library: library,

		model:    model,
		provider: provider,

		embeddingMeter: embeddingMeter,
	}
}

func (p *synthesizer) otelSetup() {
}

func (p *synthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.synthesizer.Synthesize(ctx, content, options)

	p.embeddingMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	if content != "" {
		span.SetAttributes(attribute.String("prompt", content))
	}

	return result, err
}
