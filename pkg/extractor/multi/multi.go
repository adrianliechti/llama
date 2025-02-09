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

	var content []byte

	if input.Reader != nil {
		data, err := io.ReadAll(input.Reader)

		if err != nil {
			return nil, err
		}

		content = data
	}

	for _, p := range e.providers {
		file := extractor.File{
			URL: input.URL,

			Name: input.Name,
		}

		if content != nil {
			file.Reader = bytes.NewReader(content)
		}

		result, err := p.Extract(ctx, file, options)

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
