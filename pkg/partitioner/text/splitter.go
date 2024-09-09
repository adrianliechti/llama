package text

import (
	"context"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/partitioner"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ partitioner.Provider = &Splitter{}

type Splitter struct {
	chunkSize    int
	chunkOverlap int
}

func New(options ...Option) (*Splitter, error) {
	s := &Splitter{
		chunkSize:    4000,
		chunkOverlap: 200,
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}

func (s *Splitter) Partition(ctx context.Context, input partitioner.File, options *partitioner.PartitionOptions) ([]partitioner.Partition, error) {
	if options == nil {
		options = &partitioner.PartitionOptions{}
	}

	if !isSupported(input) {
		return nil, partitioner.ErrUnsupported
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	var result []partitioner.Partition

	splitter := text.NewSplitter()
	splitter.ChunkSize = s.chunkSize
	splitter.ChunkOverlap = s.chunkOverlap

	chunks := splitter.Split(string(data))

	for i, chunk := range chunks {
		p := partitioner.Partition{
			ID:      fmt.Sprintf("%s#%d", input.Name, i),
			Content: chunk,
		}

		result = append(result, p)
	}

	return result, nil
}

func isSupported(input partitioner.File) bool {
	ext := strings.ToLower(path.Ext(input.Name))
	return slices.Contains(SupportedExtensions, ext)
}
