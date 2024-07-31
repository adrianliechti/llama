package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ObservableSynthesizer interface {
	Observable
	provider.Synthesizer
}

type observableSynthesizer struct {
	name    string
	library string

	model    string
	provider string

	synthesizer provider.Synthesizer
}

func NewSynthesizer(provider, model string, p provider.Synthesizer) ObservableSynthesizer {
	library := strings.ToLower(provider)

	return &observableSynthesizer{
		synthesizer: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-synthesizer") + "-synthesizer",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *observableSynthesizer) otelSetup() {
}

func (p *observableSynthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.synthesizer.Synthesize(ctx, content, options)

	meterRequest(ctx, p.library, p.provider, "synthesize", p.model)

	if EnableDebug {
		span.SetAttributes(attribute.String("input", content))
	}

	return result, err
}
