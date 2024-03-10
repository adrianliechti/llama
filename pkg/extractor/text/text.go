package text

import (
	"context"
	"fmt"
	"io"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ extractor.Provider = &Provider{}

type Provider struct {
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	return p, nil
}

func (p *Provider) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = &extractor.ExtractOptions{}
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	result := extractor.Document{
		Name: input.Name,
	}

	splitter := text.NewSplitter()
	splitter.ChunkSize = 4000
	splitter.ChunkOverlap = 200

	chunks := splitter.Split(string(data))

	for i, chunk := range chunks {
		block := []extractor.Block{
			{
				ID:      fmt.Sprintf("%s#%d", result.Name, i),
				Content: chunk,
			},
		}

		result.Blocks = append(result.Blocks, block...)
	}

	return &result, nil
}
