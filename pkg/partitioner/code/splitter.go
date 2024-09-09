package code

import (
	"context"
	"fmt"
	"io"
	"path"
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
		chunkSize:    1500,
		chunkOverlap: 0,
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

	splitter := text.NewSplitter()
	splitter.ChunkSize = s.chunkSize
	splitter.ChunkOverlap = s.chunkOverlap

	if sep := getSeperators(input); len(sep) > 0 {
		splitter.Separators = sep
	}

	chunks := splitter.Split(string(data))

	var result []partitioner.Partition

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
	return getSeperators(input) != nil
}

func getSeperators(input partitioner.File) []string {
	switch strings.ToLower(path.Ext(input.Name)) {
	case ".cs":
		return languageCSharp

	case ".cpp":
		return languageCPP

	case ".go":
		return languageGo

	case ".java":
		return languageJava

	case ".kt":
		return languageKotlin

	case ".js", ".jsm":
		return languageJavaScript

	case ".ts", ".tsx":
		return languageTypeScript

	case ".py":
		return languagePython

	case ".rb":
		return languageRuby

	case ".rs":
		return languageRust

	case ".sc", ".scala":
		return languageScala

	case ".swift":
		return languageSwift
	}

	return nil
}
