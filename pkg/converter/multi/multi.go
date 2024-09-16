package multi

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/converter"
)

var _ converter.Provider = &Converter{}

type Converter struct {
	providers []converter.Provider
}

func New(provider ...converter.Provider) *Converter {
	return &Converter{
		providers: provider,
	}
}

func (c *Converter) Convert(ctx context.Context, input converter.File, options *converter.ConvertOptions) (*converter.Document, error) {
	if options == nil {
		options = new(converter.ConvertOptions)
	}

	for _, p := range c.providers {
		result, err := p.Convert(ctx, input, options)

		if err != nil {
			if errors.Is(err, converter.ErrUnsupported) {
				continue
			}

			return nil, err
		}

		return result, nil
	}

	return nil, converter.ErrUnsupported
}
