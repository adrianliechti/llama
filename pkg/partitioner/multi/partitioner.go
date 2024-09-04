package multi

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/partitioner"
)

var _ partitioner.Provider = &Partitioner{}

type Partitioner struct {
	providers []partitioner.Provider
}

func New(provider ...partitioner.Provider) *Partitioner {
	return &Partitioner{
		providers: provider,
	}
}

func (p *Partitioner) Partition(ctx context.Context, input partitioner.File, options *partitioner.PartitionOptions) ([]partitioner.Partition, error) {
	if options == nil {
		options = new(partitioner.PartitionOptions)
	}

	for _, p := range p.providers {
		result, err := p.Partition(ctx, input, options)

		if err != nil {
			if errors.Is(err, partitioner.ErrUnsupported) {
				continue
			}

			return nil, err
		}

		return result, nil
	}

	return nil, partitioner.ErrUnsupported
}
