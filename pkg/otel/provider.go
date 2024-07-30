package otel

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func meterRequest(ctx context.Context, library, provider, model, kind string) {
	meter, _ := otel.Meter(library).Int64Counter("llm_requests_total")

	meter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", strings.ToLower(provider)),
		attribute.String("model", strings.ToLower(model)),
		attribute.String("kind", strings.ToLower(kind)),
	))
}
