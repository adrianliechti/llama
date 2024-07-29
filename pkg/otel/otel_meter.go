package otel

import (
	"io"
	"time"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
)

func setupMeter(resource *resource.Resource) error {
	exporter, err := stdoutmetric.New(stdoutmetric.WithWriter(io.Discard), stdoutmetric.WithPrettyPrint())

	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(3*time.Second))),
	)

	otel.SetMeterProvider(provider)

	return nil
}
