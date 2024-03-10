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
	chunkSize    int
	chunkOverlap int
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{
		chunkSize:    4000,
		chunkOverlap: 200,
	}

	for _, option := range options {
		option(p)
	}

	return p, nil
}

func WithChunkSize(size int) Option {
	return func(p *Provider) {
		p.chunkSize = size
	}
}

func WithChunkOverlap(overlap int) Option {
	return func(p *Provider) {
		p.chunkOverlap = overlap
	}
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
	splitter.ChunkSize = p.chunkSize
	splitter.ChunkOverlap = p.chunkOverlap

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
