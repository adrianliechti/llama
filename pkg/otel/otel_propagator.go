package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

func newPropagator(resource *sdkresource.Resource) error {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(propagator)

	return nil
}
