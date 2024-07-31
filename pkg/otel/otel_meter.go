package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

func setupMeter(ctx context.Context, resource *sdkresource.Resource) error {
	exporter, err := otlpmetrichttp.New(ctx)

	if err != nil {
		return err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(3*time.Second))),
		sdkmetric.WithResource(resource),
	)

	otel.SetMeterProvider(provider)

	return nil
}
