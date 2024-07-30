package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
}

func NewTranslator(provider, model string, p provider.Translator) ObservableTranslator {
	library := strings.ToLower(provider)

	return &translator{
		translator: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-translator") + "-translator",
		library: library,

		model:    model,
		provider: provider,
	}
}

func (p *translator) otelSetup() {
}

func (p *translator) Translate(ctx context.Context, content string, options *provider.TranslateOptions) (*provider.Translation, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.translator.Translate(ctx, content, options)

	meterRequest(ctx, p.library, p.provider, p.model, "translation")

	if content != "" {
		span.SetAttributes(attribute.String("input", content))
	}

	if result != nil {
		if result.Content != "" {
			span.SetAttributes(attribute.String("output", result.Content))
		}
	}

	return result, err
}
