package text

import (
	"context"
	"io"
	"path"
	"slices"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/extractor"
)

var _ extractor.Provider = &Extractor{}

type Extractor struct {
}

func New() (*Extractor, error) {
	return &Extractor{}, nil
}

func (e *Extractor) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = new(extractor.ExtractOptions)
	}

	if input.Content == nil {
		return nil, extractor.ErrUnsupported
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	if !detectText(input.Name, data) {
		return nil, extractor.ErrUnsupported
	}

	return &extractor.Document{
		Content: string(data),
	}, nil
}

func detectText(name string, data []byte) bool {
	if isSupported(name) {
		return true
	}

	var printableCount int

	for _, b := range data {
		if b == 0 {
			return false
		}

		if unicode.IsPrint(rune(b)) || b == '\n' || b == '\r' || b == '\t' {
			printableCount++
		}
	}

	return printableCount > (len(data) * 90 / 100)
}

func isSupported(name string) bool {
	ext := strings.ToLower(path.Ext(name))
	return slices.Contains(SupportedExtensions, ext)
}
