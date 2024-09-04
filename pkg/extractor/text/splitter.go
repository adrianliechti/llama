package text

import (
	"context"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ extractor.Provider = &Splitter{}

type Splitter struct {
	*Config
}

func New(options ...Option) (*Splitter, error) {
	c := &Config{
		chunkSize:    4000,
		chunkOverlap: 200,
	}

	for _, option := range options {
		option(c)
	}

	return &Splitter{
		Config: c,
	}, nil
}

func (s *Splitter) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = &extractor.ExtractOptions{}
	}

	if !isSupported(input) {
		return nil, extractor.ErrUnsupported
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	result := extractor.Document{
		Name: input.Name,
	}

	splitter := text.NewSplitter()
	splitter.ChunkSize = s.chunkSize
	splitter.ChunkOverlap = s.chunkOverlap

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

func isSupported(input extractor.File) bool {
	ext := strings.ToLower(path.Ext(input.Name))
	return slices.Contains(SupportedExtensions, ext)
}
