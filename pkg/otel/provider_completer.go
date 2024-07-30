package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type ObservableCompleter interface {
	Observable
	provider.Completer
}

type completer struct {
	name    string
	library string

	model    string
	provider string

	completer provider.Completer

	completionMeter metric.Int64Counter
}

func NewCompleter(provider, model string, p provider.Completer) ObservableCompleter {
	library := strings.ToLower(provider)

	completionMeter, _ := otel.Meter(library).Int64Counter("llm_platform_completion")

	return &completer{
		completer: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-completer") + "-completer",
		library: library,

		model:    model,
		provider: provider,

		completionMeter: completionMeter,
	}
}

func (p *completer) otelSetup() {
}

func (p *completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.completer.Complete(ctx, messages, options)

	p.completionMeter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(p.provider)),
		attribute.String("model", strings.ToLower(p.model)),
	))

	if len(messages) > 0 {
		message := messages[len(messages)-1]

		if message.Content != "" {
			span.SetAttributes(attribute.String("prompt", message.Content))
		}
	}

	if result != nil {
		if result.Message.Content != "" {
			span.SetAttributes(attribute.String("output", result.Message.Content))
		}

		if result.Reason != "" {
			span.SetAttributes(attribute.String("reason", string(result.Reason)))
		}
	}

	return result, err
}
