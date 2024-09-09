package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/partitioner"

	"go.opentelemetry.io/otel"
)

type ObservablePartitioner interface {
	Observable
	partitioner.Provider
}

type observablePartitioner struct {
	name    string
	library string

	provider string

	partitioner partitioner.Provider
}

func NewPartitioner(provider string, p partitioner.Provider) ObservablePartitioner {
	library := strings.ToLower(provider)

	return &observablePartitioner{
		partitioner: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-partitioner") + "-partitioner",
		library: library,

		provider: provider,
	}
}

func (p *observablePartitioner) otelSetup() {
}

func (p *observablePartitioner) Partition(ctx context.Context, input partitioner.File, options *partitioner.PartitionOptions) ([]partitioner.Partition, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.partitioner.Partition(ctx, input, options)

	return result, err
}
