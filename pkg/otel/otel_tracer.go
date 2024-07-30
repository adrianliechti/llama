package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func setupTracer(ctx context.Context, resource *sdkresource.Resource) error {
	exporter, err := otlptracehttp.New(ctx)

	if err != nil {
		return err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(provider)

	return nil
}
