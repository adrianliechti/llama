package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableTranscriber interface {
	Observable
	provider.Transcriber
}

type transcriber struct {
	name    string
	library string

	model    string
	provider string

	transcriber provider.Transcriber

	embeddingMeter metric.Int64Counter
}

func NewTranscriber(provider, model string, p provider.Transcriber) ObservableTranscriber {
	library := strings.ToLower(provider)

	embeddingMeter, _ := otel.Meter(library).Int64Counter("llm_platform_transcription")

	return &transcriber{
		transcriber: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-transcriber") + "-transcriber",
		library: library,

		model:    model,
		provider: provider,

		embeddingMeter: embeddingMeter,
	}
}

func (p *transcriber) otelSetup() {
}

func (p *transcriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.transcriber.Transcribe(ctx, input, options)

	p.embeddingMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	return result, err
}
