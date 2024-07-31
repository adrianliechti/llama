package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
)

type ObservableTranscriber interface {
	Observable
	provider.Transcriber
}

type observableTranscriber struct {
	name    string
	library string

	model    string
	provider string

	transcriber provider.Transcriber
}

func NewTranscriber(provider, model string, p provider.Transcriber) ObservableTranscriber {
	library := strings.ToLower(provider)

	return &observableTranscriber{
		transcriber: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-transcriber") + "-transcriber",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *observableTranscriber) otelSetup() {
}

func (p *observableTranscriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.transcriber.Transcribe(ctx, input, options)

	meterRequest(ctx, p.library, p.provider, "transcribe", p.model)

	return result, err
}
