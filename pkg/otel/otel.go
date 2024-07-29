package otel

import (
	"io"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
)

func Setup(name string) error {
	if err := newPropagator(); err != nil {
		return err
	}

	if err := setupTracer(); err != nil {
		return err
	}

	if err := setupMeter(); err != nil {
		return err
	}

	if err := setupLogger(); err != nil {
		return err
	}

	return nil
}

func newPropagator() error {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(propagator)

	return nil
}

func setupMeter() error {
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

func setupLogger() error {
	exporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())

	if err != nil {
		return err
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)

	global.SetLoggerProvider(provider)

	return nil
}
