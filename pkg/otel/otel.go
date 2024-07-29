package otel

import (
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func Setup(serviceName, serviceVersion string) error {
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)

	if err != nil {
		return err
	}

	if err := newPropagator(resource); err != nil {
		return err
	}

	if err := setupTracer(resource); err != nil {
		return err
	}

	if err := setupMeter(resource); err != nil {
		return err
	}

	if err := setupLogger(resource); err != nil {
		return err
	}

	return nil
}
