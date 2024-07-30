package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

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
	)

	otel.SetMeterProvider(provider)

	return nil
}

func Meter(name string) metric.Meter {
	return otel.Meter(name)
}

// var totalRequests = prometheus.NewCounterVec(
// 	prometheus.CounterOpts{
// 		Name: "http_requests_total",
// 		Help: "Number of get requests.",
// 	},
// 	[]string{"path"},
// )

// func ReportModel(id string) error {

// }
