package multi

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/adrianliechti/llama/pkg/extractor"
)

var _ extractor.Provider = &Extractor{}

type Extractor struct {
	providers []extractor.Provider
}

func New(provider ...extractor.Provider) *Extractor {
	return &Extractor{
		providers: provider,
	}
}

func (e *Extractor) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = new(extractor.ExtractOptions)
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	for _, p := range e.providers {
		if input.Content != nil {
			input.Content = bytes.NewReader(data)
		}

		result, err := p.Extract(ctx, input, options)

		if err != nil {
			if errors.Is(err, extractor.ErrUnsupported) {
				continue
			}

			return nil, err
		}

		return result, nil
	}

	return nil, extractor.ErrUnsupported
}
