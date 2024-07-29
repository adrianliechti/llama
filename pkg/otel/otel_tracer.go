package otel

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer(name).Start(ctx, name)
}

func HTTPClient() *http.Client {
	return &http.Client{
		Transport: HTTPTransport(),
	}
}

func HTTPTransport() http.RoundTripper {
	return HTTPTransportWith(nil)
}

func HTTPTransportWith(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return otelhttp.NewTransport(rt)
}
