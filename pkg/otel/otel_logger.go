package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/log/global"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

func setupLogger(ctx context.Context, resource *sdkresource.Resource) error {
	exporter, err := otlploghttp.New(ctx)

	if err != nil {
		return err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(resource),
	)

	global.SetLoggerProvider(provider)

	return nil
}

func Logger(name string) *slog.Logger {
	provider := global.GetLoggerProvider()
	return otelslog.NewLogger(name, otelslog.WithLoggerProvider(provider))
}
