package otel

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func meterRequest(ctx context.Context, library, provider, operation, model string) {
	meter, _ := otel.Meter(library).Int64Counter("llm_requests_total")

	meter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(provider)),
		attribute.String("operation", strings.ToLower(operation)),
		attribute.String("model", strings.ToLower(model)),
	))
}

func meterTokens(ctx context.Context, library, provider, operation, model string, val int64) {
	meter, _ := otel.Meter(library).Int64Counter("llm_tokens_total")

	meter.Add(ctx, val, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(provider)),
		attribute.String("operation", strings.ToLower(operation)),
		attribute.String("model", strings.ToLower(model)),
	))
}
