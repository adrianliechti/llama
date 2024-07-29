package otel

import (
	"go.opentelemetry.io/otel/log/global"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
)

func setupLogger(resource *resource.Resource) error {
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
